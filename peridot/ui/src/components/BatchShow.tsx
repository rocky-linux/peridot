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
import { Header } from 'common/mui/Header';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import Divider from '@mui/material/Divider';
import TabContext from '@mui/lab/TabContext';
import TabList from '@mui/lab/TabList';
import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { useParams } from 'react-router';
import { transformTaskStatusToIcon } from 'peridot/ui/src/components/ProjectTasks';
import {
  V1TaskStatus,
  GetImportBatchRequest,
  GetBuildBatchRequest,
} from 'bazel-bin/peridot/proto/v1/client_typescript';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import { taskApi } from 'peridot/ui/src/api';
import ProjectTaskFullLog from 'peridot/ui/src/components/ProjectTaskFullLog';
import AppBar from '@mui/material/AppBar';
import Tabs from '@mui/material/Tabs';
import { BuildsTable } from 'peridot/ui/src/components/BuildsTable';
import { ImportsTable } from 'peridot/ui/src/components/ImportsTable';

export interface BatchShowParams {
  id: string;
}

export interface BatchShowProps<T> {
  name: string;
  pluralName: string;
  fetchBatch: (req: T) => any;
}

export const BatchShow = <
  T extends GetImportBatchRequest | GetBuildBatchRequest
>(
  props: BatchShowProps<T>
) => {
  const project = React.useContext(ProjectContext);

  if (!project) {
    return null;
  }

  const params = useParams<BatchShowParams>();
  const [resResource, setResResource] = React.useState<any>();
  const [resState, setResState] = React.useState<any>();
  const [tabValue, setTabValue] = React.useState('1');
  const [pageSize, setPageSize] = React.useState(100);
  const [page, setPage] = React.useState(0);

  const handleTabValueChange = (event, newValue) => {
    setTabValue(newValue);
  };

  const idName = `${props.name}BatchId`;

  fetchRemoteResource(() => {
    const req = {
      projectId: project.id || '',
      page: 0,
      limit: 1,
    } as any;
    req[idName] = params.id;

    return props.fetchBatch(req);
  }, setResResource);

  fetchRemoteResource(
    () => {
      const req = {
        projectId: project.id || '',
        limit: pageSize,
        filterStatus: tabValue,
        page,
      } as any;
      req[idName] = params.id;

      return props.fetchBatch(req);
    },
    setResState,
    false,
    [tabValue, pageSize, page]
  );

  React.useEffect(() => {
    if (resResource) {
      let tabVal = '1';
      if (resResource.running !== 0 && resResource.running) {
        tabVal = '2';
      }
      if (resResource.succeeded !== 0 && resResource.succeeded) {
        tabVal = '3';
      }
      if (resResource.failed !== 0 && resResource.failed) {
        tabVal = '4';
      }
      if (resResource.canceled !== 0 && resResource.canceled) {
        tabVal = '5';
      }
      if (tabVal !== tabValue) {
        setTabValue(tabVal);
      }
    }
  }, [resResource]);

  return (
    <>
      {suspenseRemoteResource(resResource, (res) => (
        <>
          <ToolbarHeader>
            <div className="flex space-x-4 items-center">
              <div className="flex flex-col">
                <h5 className="font-bold text-xs">Project</h5>
                <h5 className="font-light text-xs">{project.name}</h5>
              </div>
              <ChevronRightIcon />
              <div className="flex flex-col">
                <h5 className="font-bold text-xs">
                  {props.name.substr(0, 1).toUpperCase() + props.name.substr(1)}{' '}
                  Batch ID
                </h5>
                <h5 className="font-light text-xs">{params.id}</h5>
              </div>
            </div>
          </ToolbarHeader>
          <Divider />
          <div className="w-full bg-white flex divide-x divide-gray-100">
            {res.pending !== 0 && (
              <>
                <div className="p-5 w-128 space-y-1 space-x-4">
                  <h2 className="font-bold text-lg">
                    {transformTaskStatusToIcon(
                      V1TaskStatus.Pending,
                      true,
                      res.pending
                    )}
                  </h2>
                </div>
              </>
            )}
            {res.running !== 0 && (
              <>
                <div className="p-5 w-128 space-y-1 space-x-4">
                  <h2 className="font-bold text-lg">
                    {transformTaskStatusToIcon(
                      V1TaskStatus.Running,
                      true,
                      res.running
                    )}
                  </h2>
                </div>
              </>
            )}
            {res.succeeded !== 0 && (
              <>
                <div className="p-5 w-128 space-y-1 space-x-4">
                  <h2 className="font-bold text-lg">
                    {transformTaskStatusToIcon(
                      V1TaskStatus.Succeeded,
                      true,
                      res.succeeded
                    )}
                  </h2>
                </div>
              </>
            )}
            {res.failed !== 0 && (
              <>
                <div className="p-5 w-128 space-y-1 space-x-4">
                  <h2 className="font-bold text-lg">
                    {transformTaskStatusToIcon(
                      V1TaskStatus.Failed,
                      true,
                      res.failed
                    )}
                  </h2>
                </div>
              </>
            )}
            {res.canceled !== 0 && (
              <>
                <div className="p-5 w-128 space-y-1 space-x-4">
                  <h2 className="font-bold text-lg">
                    {transformTaskStatusToIcon(
                      V1TaskStatus.Canceled,
                      true,
                      res.canceled
                    )}
                  </h2>
                </div>
              </>
            )}
          </div>
          <Divider />
          <Box className="bg-white">
            <Tabs value={tabValue} onChange={handleTabValueChange}>
              {res.pending !== 0 && <Tab label="Pending" value="1" />}
              {res.running !== 0 && <Tab label="Running" value="2" />}
              {res.succeeded !== 0 && <Tab label="Succeeded" value="3" />}
              {res.failed !== 0 && <Tab label="Failed" value="4" />}
              {res.canceled !== 0 && <Tab label="Canceled" value="5" />}
            </Tabs>
          </Box>
          {suspenseRemoteResource(resState, (res2) => (
            <>
              {props.name === 'build' && (
                <BuildsTable
                  builds={res2.builds}
                  total={res2.total}
                  page={page}
                  pageSize={pageSize}
                  setPage={setPage}
                  setPageSize={setPageSize}
                />
              )}
              {props.name === 'import' && (
                <ImportsTable
                  imports={res2.imports}
                  total={res2.total}
                  page={page}
                  pageSize={pageSize}
                  setPage={setPage}
                  setPageSize={setPageSize}
                />
              )}
            </>
          ))}
        </>
      ))}
    </>
  );
};
