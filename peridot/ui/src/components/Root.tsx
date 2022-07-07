/*
 * Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
 * Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
 * Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 * this list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 * this list of conditions and the following disclaimer in the documentation
 * and/or other materials provided with the distribution.
 *
 * 3. Neither the name of the copyright holder nor the names of its contributors
 * may be used to endorse or promote products derived from this software without
 * specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

import React from 'react';
import { Route, RouteComponentProps, Switch } from 'react-router';
import * as H from 'history';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import ImportExportIcon from '@mui/icons-material/ImportExport';
import BuildIcon from '@mui/icons-material/Build';
import BuildCircleIcon from '@mui/icons-material/BuildCircle';
import ArrowCircleDownIcon from '@mui/icons-material/ArrowCircleDown';
import TaskAltIcon from '@mui/icons-material/TaskAlt';
import ListAltIcon from '@mui/icons-material/ListAlt';
import SettingsIcon from '@mui/icons-material/Settings';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import DashboardIcon from '@mui/icons-material/Dashboard';
import CloudSyncIcon from '@mui/icons-material/CloudSync';
import ImagesearchRollerIcon from '@mui/icons-material/ImagesearchRoller';
import CategoryIcon from '@mui/icons-material/Category';
import PasswordIcon from '@mui/icons-material/Password';
import {
  fetchRemoteResource,
  setLoadingElement,
  suspenseRemoteResource,
} from 'common/ui/remote';
import { PageWrapper } from 'dotui/PageWrapper';
import { RippleLoading } from 'dotui/RippleLoading';
import { Navbar } from 'dotui';
import { PeridotLogo } from 'common/ui/PeridotLogo';
import { generateSuspensedComponent } from 'common/ui/SuspensedComponent';
import { V1GetProjectResponse } from 'bazel-bin/peridot/proto/v1/client_typescript';
import { projectApi } from '../api';
import { ProjectContext } from '../context/ProjectContext';
import { NavbarDrawer } from 'common/mui/NavbarDrawer';
import Button from '@mui/material/Button';
import { RemoteErrors } from 'common/ui/types';
import Alert from '@mui/material/Alert';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import Divider from '@mui/material/Divider';
import NotFound from './NotFound';

setLoadingElement(
  <>
    <RippleLoading />
  </>
);

export const Root = () => {
  const pathname = location.pathname.split('/');

  const [projectRes, setProjectRes] = React.useState<
    V1GetProjectResponse | RemoteErrors | undefined | null
  >(undefined);
  const [projectId, setProjectId] = React.useState<string | undefined>(
    pathname.length >= 2
      ? pathname[1].substring(0, 1) === '_'
        ? undefined
        : pathname[1]
      : undefined
  );

  React.useEffect(() => {
    if (projectId) {
      fetchRemoteResource(
        () => projectApi.getProject({ id: projectId }),
        setProjectRes,
        true
      );
    }
  }, []);

  const indexActive = (_: unknown, location: H.Location): boolean => {
    const pathname = location.pathname;
    const pathnameSplit = pathname.split('/');

    return (
      (pathnameSplit.length === 2 && pathnameSplit[1] === '') ||
      (pathnameSplit.length === 3 && pathnameSplit[2] === '')
    );
  };

  return (
    <ProjectContext.Provider
      value={
        projectRes === null || projectRes === 'access_denied'
          ? null
          : projectRes?.project
      }
    >
      <div className="flex w-full min-h-screen">
        <NavbarDrawer
          logo={(classes: string) => (
            <div className="text-2xl text-center w-full">Peridot</div>
          )}
          afterLogoNode={
            projectRes !== 'access_denied' &&
            projectRes?.project &&
            (() => (
              <div>
                <Button variant="outlined" sx={{ color: 'white' }} href="/">
                  <div className="flex items-center space-x-1">
                    <p className="w-36 text-xs text-ellipsis overflow-hidden whitespace-nowrap">
                      {projectRes?.project?.name}
                    </p>
                    <ArrowDropDownIcon />
                  </div>
                </Button>
              </div>
            ))
          }
          mainLinks={
            projectRes === 'access_denied' || !projectRes?.project
              ? [
                  {
                    links: [
                      {
                        text: 'Projects',
                        href: '/',
                        real: true,
                        icon: (classes) => (
                          <AccountTreeIcon className={classes} />
                        ),
                      },
                    ],
                  },
                ]
              : [
                  {
                    links: [
                      {
                        text: 'Overview',
                        href: '/',
                        icon: (classes: string) => (
                          <DashboardIcon className={classes} />
                        ),
                        isActive: indexActive,
                      },
                      {
                        text: 'Imports',
                        href: '/imports',
                        icon: (classes) => (
                          <ImportExportIcon className={classes} />
                        ),
                      },
                      {
                        text: 'Builds',
                        href: '/builds',
                        icon: (classes) => <BuildIcon className={classes} />,
                      },
                      {
                        text: 'Import batches',
                        href: '/import_batches',
                        icon: (classes) => (
                          <ArrowCircleDownIcon className={classes} />
                        ),
                      },
                      {
                        text: 'Build batches',
                        href: '/build_batches',
                        icon: (classes) => (
                          <BuildCircleIcon className={classes} />
                        ),
                      },
                      {
                        text: 'Tasks',
                        href: '/tasks',
                        icon: (classes) => <TaskAltIcon className={classes} />,
                      },
                    ],
                  },
                  {
                    title: 'Metadata',
                    links: [
                      {
                        text: 'Packages',
                        href: '/packages',
                        icon: (classes) => <ListAltIcon className={classes} />,
                      },
                      window.state.email && {
                        text: 'Settings',
                        href: '/settings',
                        icon: (classes) => <SettingsIcon className={classes} />,
                      },
                      window.state.email && {
                        text: 'Credentials',
                        href: '/credentials',
                        icon: (classes) => <PasswordIcon className={classes} />,
                      },
                      {
                        text: 'Repositories',
                        href: '/repositories',
                        icon: (classes) => <CategoryIcon className={classes} />,
                      },
                      window.state.email && {
                        text: 'Sync',
                        href: '/sync',
                        icon: (classes) => (
                          <CloudSyncIcon className={classes} />
                        ),
                      },
                    ],
                  },
                ]
          }
        />
        <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
          <Box className="flex flex-col" component="main" sx={{ flex: 1, bgcolor: '#eaeff1', height: '100%' }}>
            {projectRes === 'access_denied' ? (
              <PageWrapper>
                <Alert severity="error">
                  You don't have access to this project
                </Alert>
              </PageWrapper>
            ) : (
              <Switch>
                {!projectId && (
                  <>
                    <Route
                      path="/"
                      exact
                      component={() => {
                        if (projectRes?.project) {
                          setProjectRes(undefined);
                          setProjectId(undefined);
                        }

                        return (
                          <>
                            {generateSuspensedComponent(
                              () => import('./ProjectList')
                            )(undefined)}
                          </>
                        );
                      }}
                    />
                    <Route
                      path="/_projects/new"
                      exact
                      component={generateSuspensedComponent(
                        () => import('./ProjectCreate')
                      )}
                    />
                  </>
                )}
                {projectId && (
                  <>
                    <Route
                      path="/"
                      exact
                      component={generateSuspensedComponent(
                        () => import('./ProjectRoot')
                      )}
                    />
                    <Route
                      path="/:projectId"
                      component={generateSuspensedComponent(
                        () => import('./ProjectWrapper')
                      )}
                    />
                  </>
                )}
                <Route component={NotFound} />
              </Switch>
            )}
          </Box>
        </Box>
      </div>
    </ProjectContext.Provider>
  );
};
