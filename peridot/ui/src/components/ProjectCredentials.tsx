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
import { Header } from 'common/mui/Header';
import { ProjectForm } from 'peridot/ui/src/components/forms/ProjectForm';
import { PageWrapper } from 'dotui/PageWrapper';
import Paper from '@mui/material/Paper';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { GitlabCredentialsForm } from 'peridot/ui/src/components/forms/GitlabCredentialsForm';
import { RemoteErrors } from 'common/ui/types';
import { V1GetProjectCredentialsResponse } from 'bazel-bin/peridot/proto/v1/client_typescript';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import { packageApi, projectApi } from 'peridot/ui/src/api';

export default function () {
  const project = React.useContext(ProjectContext);

  const [credentialsRes, setCredentialsRes] = React.useState<
    V1GetProjectCredentialsResponse | RemoteErrors | undefined | null
  >(undefined);

  fetchRemoteResource(
    () =>
      projectApi.getProjectCredentials({
        projectId: project?.id || '',
      }),
    setCredentialsRes
  );

  return (
    <>
      <Header title="Credentials" />
      <PageWrapper>
        <Paper
          className="max-w-5xl"
          sx={{ margin: 'auto', overflow: 'hidden' }}
        >
          <AppBar position="static" color="default" elevation={0}>
            <Toolbar>
              <p>Gitlab</p>
            </Toolbar>
          </AppBar>
          <div className="m-4 space-y-4">
            {suspenseRemoteResource(credentialsRes, (res) => (
              <GitlabCredentialsForm username={res?.gitlabUsername} />
            ))}
          </div>
        </Paper>
      </PageWrapper>
    </>
  );
}
