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
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import { DenseList } from 'common/mui/DenseList';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import Divider from '@mui/material/Divider';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { projectApi, ResourceShowParams } from 'peridot/ui/src/api';
import { useParams } from 'react-router';
import { RemoteErrors } from 'common/ui/types';
import { V1GetRepositoryResponse } from 'bazel-bin/peridot/proto/v1/client_typescript';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';

export default function () {
  const project = React.useContext(ProjectContext);
  if (!project) {
    return null;
  }

  const params = useParams<ResourceShowParams>();
  const [repoRes, setRepoRes] = React.useState<
    V1GetRepositoryResponse | RemoteErrors | undefined | null
  >(undefined);

  fetchRemoteResource(
    () =>
      projectApi.getRepository({
        projectId: project?.id || '',
        id: params.id,
      }),
    setRepoRes
  );

  return (
    <>
      {suspenseRemoteResource(repoRes, (res) => (
        <>
          <ToolbarHeader>
            <div className="flex space-x-4 items-center">
              <div className="flex flex-col">
                <h5 className="font-bold text-xs">Project</h5>
                <h5 className="font-light text-xs">{project.name}</h5>
              </div>
              <ChevronRightIcon />
              <div className="flex flex-col">
                <h5 className="font-bold text-xs">Repository</h5>
                <h5 className="font-light text-xs">{res.repository?.name}</h5>
              </div>
            </div>
          </ToolbarHeader>
          <Divider />
          <Box className="bg-white">
            <Tabs value="0">
              <Tab label="Details" value="0" />
            </Tabs>
          </Box>
          <Divider />
          <div className="p-6 bg-white flex justify-between items-start">
            <DenseList
              entries={[
                {
                  key: 'Packages',
                  entries:
                    (res.repository?.packages?.length ?? 0) > 0
                      ? res.repository?.packages?.map((pkg) => ({
                          value: pkg,
                        }))
                      : [{ value: 'No packages' }],
                },
              ]}
            />
            <DenseList
              entries={[
                {
                  key: 'Exclude filter',
                  entries:
                    (res.repository?.excludeFilter?.length ?? 0) > 0
                      ? res.repository?.excludeFilter?.map((pkg) => ({
                          value: pkg,
                        }))
                      : [{ value: 'No exclusions' }],
                },
              ]}
            />
            <DenseList
              entries={[
                {
                  key: 'Include list',
                  entries:
                    (res.repository?.includeList?.length ?? 0) > 0
                      ? res.repository?.includeList?.map((pkg) => ({
                          value: pkg,
                        }))
                      : [{ value: 'No include entries' }],
                },
              ]}
            />
          </div>
        </>
      ))}
    </>
  );
}
