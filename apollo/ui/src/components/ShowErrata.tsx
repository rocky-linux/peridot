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
  V1Advisory,
  V1GetAdvisoryResponse,
} from 'bazel-bin/apollo/proto/v1/client_typescript';
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
  Typography,
} from '@mui/material';
import { severityToBadge, severityToText, typeToText } from 'apollo/ui/src/enumToText';

interface ShowErrataParams {
  id: string;
}

export interface ShowErrataProps
  extends RouteComponentProps<ShowErrataParams> {}

export const ShowErrata = (props: ShowErrataProps) => {
  const [errata, setErrata] = React.useState<
    V1Advisory | undefined | null
  >();
  const [tabValue, setTabValue] = React.useState(0);

  React.useEffect(() => {
    (async () => {
      let err, res: void | V1GetAdvisoryResponse | undefined;
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
        <Typography variant="h2" className="text-lg text-red-800 font-bold">
          Oh no! Something has gone wrong!
        </Typography>
      )}
      {errata && (
        <>
          <div className="flex items-center justify-between mt-4 mb-6">
            <Typography variant="h5">{errata.name} </Typography>
            <div className="flex space-x-4 h-full">
              {severityToBadge(errata.severity)}
              <Chip
                color="primary"
                label={`Issued at ${errata.publishedAt?.toLocaleDateString()}`}
              />
            </div>
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
              <CardContent className="max-w-5xl space-y-4">
                <div>
                  <Typography variant="h6">Synopsis</Typography>
                  {errata.synopsis}
                </div>
                <div>
                  <Typography variant="h6">Type</Typography>
                  {typeToText(errata.type)}
                </div>
                <div>
                  <Typography variant="h6">Severity</Typography>
                  {severityToText(errata.severity)}
                </div>
                <div>
                  <Typography variant="h6">Topic</Typography>
                  {errata.topic?.split('\n').map((x) => (
                    <p>{x}</p>
                  ))}
                </div>
                <div>
                  <Typography variant="h6">Description</Typography>
                  {errata.description?.split('\n').map((x) => (
                    <p>{x}</p>
                  ))}
                </div>
                <div>
                  <Typography variant="h6">Affected products</Typography>
                  <ul>
                    {errata.affectedProducts?.map((x) => (
                      <li>{x}</li>
                    ))}
                  </ul>
                </div>
                <div>
                  <Typography variant="h6">Fixes</Typography>
                  <ul>
                    {errata.fixes?.map((x) => (
                      <li>
                        <a
                          href={x.sourceLink}
                          target="_blank"
                        >
                          {x.sourceBy} - {x.ticket}
                        </a>
                      </li>
                    ))}
                  </ul>
                </div>
                <div>
                  <Typography variant="h6">CVEs</Typography>
                  <ul>
                    {errata.cves?.map((x) => {
                      let text = `${x.name}${
                        x.sourceBy !== '' && ` (Source: ${x.sourceBy})`
                      }`;

                      return (
                        <li>
                          {x.sourceLink === '' ? (
                            <span>{text}</span>
                          ) : (
                            <a href={x.sourceLink} target="_blank">
                              {text}
                            </a>
                          )}
                        </li>
                      );
                    })}
                    {errata.cves?.length === 0 && <li>No CVEs</li>}
                  </ul>
                </div>
                <div>
                  <Typography variant="h6">References</Typography>
                  <ul>
                    {errata.references?.map((x) => (
                      <li>{x}</li>
                    ))}
                    {errata.references?.length === 0 && <li>No references</li>}
                  </ul>
                </div>
              </CardContent>
            )}
            {tabValue === 1 && (
              <CardContent className="max-w-5xl">
                <div className="space-x-4 divide-y py-2">
                  {Object.keys(errata.rpms || {}).map(product => (
                    <div className="space-y-4">
                      <Typography variant="h6">{product}</Typography>
                      <div>
                        <Typography variant="subtitle1">SRPMs</Typography>
                        <ul>
                          {errata.rpms[product].nvras
                          ?.filter((x) => x.indexOf('.src.rpm') !== -1)
                          .map((x) => (
                            <li>{x}</li>
                          ))}
                        </ul>
                      </div>
                      <div>
                        <Typography variant="subtitle1">RPMs</Typography>
                        <ul>
                          {errata.rpms[product].nvras
                          ?.filter((x) => x.indexOf('.src.rpm') === -1)
                          .map((x) => (
                            <li>{x}</li>
                          ))}
                        </ul>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            )}
          </Card>
        </>
      )}
    </div>
  );
};
