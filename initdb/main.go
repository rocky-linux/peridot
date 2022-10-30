// Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
// Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
// Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"peridot.resf.org/utils"
)

var root = &cobra.Command{
	Use: "initdb",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

func init() {
	cnf.DefaultPort = 9999

	dname := "initdb"
	cnf.DatabaseName = &dname
	cnf.Name = "initdb"

	pf := root.PersistentFlags()
	pf.String("target.db", "", "target db to initialize")
	pf.Bool("skip", false, "Whether to skip InitDB without removing it as an init container")

	utils.AddFlags(pf, cnf)
}

func mn(_ *cobra.Command, _ []string) {
	if viper.GetBool("skip") {
		os.Exit(0)
	}

	ctx := context.TODO()

	targetDB := viper.GetString("target.db")
	if targetDB == "" {
		log.Fatal("no target db")
	}

	env := os.Getenv("RESF_ENV")
	namespace := os.Getenv("RESF_NS")
	roleName := fmt.Sprintf("%s-%s", namespace, targetDB)
	secretName := fmt.Sprintf("%s-database-password", targetDB)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	_, err = clientset.CoreV1().Secrets(namespace).Get(ctx, "env", metav1.GetOptions{})
	if err != nil {
		_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "env",
				Namespace: namespace,
			},
			StringData: map[string]string{
				"hydra": uuid.New().String(),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}

	check, err := clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err == nil || check.Data["password"] != nil || len(check.Data["password"]) > 0 {
		os.Exit(0)
	}

	secret, err := clientset.CoreV1().Secrets(fmt.Sprintf("initdb-%s", env)).Get(ctx, "initdb-password", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	newUrl := strings.Replace(viper.GetString("database.url"), "REPLACEME", url.QueryEscape(string(secret.Data["password"])[:]), 1)
	if secret.Data["username"] != nil {
		newUrl = strings.Replace(newUrl, "postgres:", fmt.Sprintf("%s:", url.QueryEscape(string(secret.Data["username"]))), 1)
	}
	viper.Set("database.url", newUrl)

	pg := utils.PgInit()

	err = func() error {
		pw := uuid.New().String()
		_, err := pg.Exec(fmt.Sprintf("create database \"%s\"", targetDB))
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}

		_, err = pg.Exec(fmt.Sprintf("create role \"%s\";", roleName))
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}

		_, err = pg.Exec(fmt.Sprintf("alter role \"%s\" with password '%s'; alter role \"%s\" with login; grant all on database \"%s\" to \"%s\"", roleName, pw, roleName, targetDB, roleName))
		if err != nil {
			return err
		}

		dbUrl := viper.GetString("database.url")
		dbUrl = strings.Replace(dbUrl, "/postgres?", "/REPLACEDB?", 1)
		dbUrl = strings.Replace(dbUrl, "/initdb?", "/REPLACEDB?", 1)
		dbUrl = strings.Replace(dbUrl, "/REPLACEDB?", fmt.Sprintf("/%s?", targetDB), 1)
		viper.Set("database.url", dbUrl)
		pgInDb := utils.PgInit()
		_, err = pgInDb.Exec(fmt.Sprintf("grant all privileges on all tables in schema public to \"%s\"; grant all privileges on all sequences in schema public to \"%s\";", roleName, roleName))
		if err != nil {
			return err
		}

		_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			StringData: map[string]string{
				"password": pw,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
