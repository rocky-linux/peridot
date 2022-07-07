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
import {
  SecparseAdvisory,
  SecparseGetAdvisoryResponse,
} from 'bazel-bin/secparse/proto/v1/client_typescript';
import { reqap } from 'common/ui/reqap';
import { api } from '../api';
import { RouteComponentProps } from 'react-router';
import {
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Paper,
  Tab,
  Tabs,
} from '@material-ui/core';

interface ShowErrataParams {
  id: string;
}

export interface ShowErrataProps
  extends RouteComponentProps<ShowErrataParams> {}

export const ShowErrata = (props: ShowErrataProps) => {
  const [errata, setErrata] = React.useState<
    SecparseAdvisory | undefined | null
  >();
  const [tabValue, setTabValue] = React.useState(0);

  React.useEffect(() => {
    (async () => {
      let err, res: void | SecparseGetAdvisoryResponse | undefined;
      [err, res] = await reqap(() =>
        api.getAdvisory({ id: props.match.params.id })
      );
      if (err || !res) {
        setErrata(null);
        return;
      }

      if (res) {
        setErrata(res.advisory);
      }
    })().then();
  }, []);

  const handleTabChange = ({}, val: number) => {
    setTabValue(val);
  };

  return (
    <div>
      {errata === undefined && <CircularProgress />}
      {errata === null && (
        <h2 className="text-lg text-red-800 font-bold">
          Oh no! Something has gone wrong!
        </h2>
      )}
      {errata && (
        <>
          <div className="flex items-center justify-between">
            <h1>{errata.name} </h1>
            <Chip
              color="primary"
              label={`Issued at ${errata.publishedAt?.toLocaleDateString()}`}
            />
          </div>
          <Card>
            <Paper square>
              <Tabs
                value={tabValue}
                indicatorColor="primary"
                textColor="primary"
                onChange={handleTabChange}
                aria-label="disabled tabs example"
              >
                <Tab value={0} label="Erratum" />
                <Tab value={1} label="Affected packages" />
              </Tabs>
            </Paper>
            {tabValue === 0 && (
              <CardContent className="max-w-5xl">
                <h3>Synopsis</h3>
                {errata.synopsis}
                <h3>Type</h3>
                {errata.type}
                <h3>Severity</h3>
                {errata.severity}
                <h3>Topic</h3>
                {errata.topic?.split('\n').map((x) => (
                  <p>{x}</p>
                ))}
                <h3>Description</h3>
                {errata.description?.split('\n').map((x) => (
                  <p>{x}</p>
                ))}
                <h3>Affected products</h3>
                <ul>
                  {errata.affectedProducts?.map((x) => (
                    <li>{x}</li>
                  ))}
                </ul>
                <h3>Fixes</h3>
                <ul>
                  {errata.fixes?.map((x) => (
                    <li>
                      <a
                        href={`https://bugzilla.redhat.com/show_bug.cgi?id=${x}`}
                        target="_blank"
                      >
                        Red Hat BZ - {x}
                      </a>
                    </li>
                  ))}
                </ul>
                <h3>CVEs</h3>
                <ul>
                  {errata.cves?.map((x) => {
                    const cve = x.split(':::');
                    let text = `${cve[2]}${
                      cve[0] !== '' && ` (Source: ${cve[0]})`
                    }`;

                    return (
                      <li>
                        {cve[1] === '' ? (
                          <span>{text}</span>
                        ) : (
                          <a href={cve[1]} target="_blank">
                            {text}
                          </a>
                        )}
                      </li>
                    );
                  })}
                  {errata.cves?.length === 0 && <li>No CVEs</li>}
                </ul>
                <h3>References</h3>
                <ul>
                  {errata.references?.map((x) => (
                    <li>{x}</li>
                  ))}
                  {errata.references?.length === 0 && <li>No references</li>}
                </ul>
              </CardContent>
            )}
            {tabValue === 1 && (
              <CardContent className="max-w-5xl">
                <h3>SRPMs</h3>
                <ul>
                  {errata.rpms
                    ?.filter((x) => x.indexOf('.src.rpm') !== -1)
                    .map((x) => (
                      <li>{x}</li>
                    ))}
                </ul>
                <h3>RPMs</h3>
                <ul>
                  {errata.rpms
                    ?.filter((x) => x.indexOf('.src.rpm') === -1)
                    .map((x) => (
                      <li>{x}</li>
                    ))}
                </ul>
              </CardContent>
            )}
          </Card>
        </>
      )}
    </div>
  );
};
