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
import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import TabContext from '@mui/lab/TabContext';
import TabList from '@mui/lab/TabList';
import TabPanel from '@mui/lab/TabPanel';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import { PageWrapper } from 'dotui/PageWrapper';
import { H1 } from 'dotui/H1';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import { packageApi } from 'peridot/ui/src/api';
import {
  V1ListPackagesResponse,
  V1Package,
  V1GetPackageResponse,
} from 'bazel-bin/peridot/proto/v1/client_typescript';
import { RouteComponentProps } from 'react-router';
import { Table, TableCol, TableRow } from 'dotui/Table';
import { PeridotLink } from 'common/ui/PeridotLink';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import Divider from '@mui/material/Divider';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import { generateSuspensedComponent } from 'common/ui/SuspensedComponent';
import { RemoteErrors } from 'common/ui/types';

export interface ProjectPackageDetailsRouteProps {
  name: string;
}

export default function (
  props: RouteComponentProps<ProjectPackageDetailsRouteProps>
) {
  const project = React.useContext(ProjectContext);
  if (!project) {
    return null;
  }

  const [packageRes, setPackageRes] = React.useState<
    V1GetPackageResponse | undefined | null | RemoteErrors
  >(undefined);
  const [value, setValue] = React.useState('1');

  const handleChange = (event, newValue) => {
    setValue(newValue);
  };

  fetchRemoteResource(
    () =>
      packageApi.getPackage({
        projectId: project?.id || '',
        field: 'name',
        value: props.match.params.name,
      }),
    setPackageRes
  );

  return (
    <>
      {suspenseRemoteResource(packageRes, (res: V1GetPackageResponse) => {
        const pkg = (res as any)['package'];
        return (
          pkg && (
            <>
              <ToolbarHeader>
                <div className="flex space-x-4 items-center">
                  <div className="flex flex-col">
                    <h5 className="font-bold text-xs">Project</h5>
                    <h5 className="font-light text-xs">{project.name}</h5>
                  </div>
                  <ChevronRightIcon />
                  <div className="flex flex-col">
                    <h5 className="font-bold text-xs">Package</h5>
                    <h5 className="font-light text-xs">{pkg.name}</h5>
                  </div>
                </div>
              </ToolbarHeader>
              <Divider />
              <div className="w-full bg-white flex divide-x divide-gray-100">
                <div className="p-5 w-64 space-y-1">
                  <h2 className="font-bold text-lg">Active version</h2>
                  <div className="h-6">4.4.19.14.el8_3</div>
                </div>
                <div className="p-5 w-128 space-y-1">
                  <h2 className="font-bold text-lg">Import</h2>
                  <div className="flex space-x-2 h-6">
                    <CheckCircleIcon className="text-green-500" />
                    <PeridotLink to="/imports/b7775543-44a1-4d17-bbc4-e6873f8476b0">
                      b7775543-44a1-4d17-bbc4-e6873f8476b0
                    </PeridotLink>
                  </div>
                </div>
                <div className="p-5 w-128 space-y-1">
                  <h2 className="font-bold text-lg">Latest build</h2>
                  <div className="flex space-x-2 h-6">
                    <CheckCircleIcon className="text-green-500" />
                    <PeridotLink to="/builds/a6050c79-f350-48df-914a-045da5e44092">
                      a6050c79-f350-48df-914a-045da5e44092
                    </PeridotLink>
                  </div>
                </div>
              </div>
              <Divider />
              <TabContext value={value}>
                <Box className="bg-white">
                  <TabList onChange={handleChange}>
                    <Tab label="Builds" value="1" />
                    <Tab label="Imports" value="2" />
                    <Tab label="Artifacts" value="3" />
                    <Tab label="Other projects" value="4" />
                  </TabList>
                </Box>
                <TabPanel value="1">
                  <div className="m-[-24px]">
                    {React.createElement(
                      generateSuspensedComponent(
                        () => import('./ProjectPackageBuildsTab')
                      )
                    )}
                  </div>
                </TabPanel>
                <TabPanel value="2">Item Two</TabPanel>
                <TabPanel value="3">Item Three</TabPanel>
                <TabPanel value="4">Item Four</TabPanel>
              </TabContext>
            </>
          )
        );
      })}
    </>
  );
}
