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
  V1ListProjectsResponse,
  V1Project,
} from 'bazel-bin/peridot/proto/v1/client_typescript';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import { RemoteState } from 'common/ui/types';
import { projectApi } from '../api';
import { PageWrapper } from 'dotui/PageWrapper';
import { H1 } from 'dotui/H1';
import { Card } from 'dotui/Card';
import { H5 } from 'dotui/H5';
import { H3 } from 'dotui/H3';
import { Link } from 'react-router-dom';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import { PeridotLink } from 'common/ui/PeridotLink';
import { GridColDef, DataGrid } from '@mui/x-data-grid';
import { Header } from 'common/mui/Header';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';

const columns: GridColDef[] = [
  {
    field: 'id',
    headerName: 'ID',
    sortable: false,
    flex: 1,
    renderCell: (params) => (
      <PeridotLink real to={`/${params.value}`}>
        {params.value}
      </PeridotLink>
    ),
  },
  {
    field: 'name',
    headerName: 'Name',
    sortable: false,
    flex: 1,
  },
  {
    field: 'distTag',
    headerName: 'Dist',
    sortable: false,
    flex: 1,
  },
];

export default function () {
  const [projects, setProjects] =
    React.useState<RemoteState<V1ListProjectsResponse>>(undefined);

  fetchRemoteResource(() => projectApi.listProjects({}), setProjects);

  const toolbarChildren = (
    <>
      <Grid item>
        <Link to="/_projects/new">
          <Button variant="outlined" color="inherit" size="small">
            New project
          </Button>
        </Link>
      </Grid>
    </>
  );

  return (
    <>
      <Header title="Projects" toolbarChildren={toolbarChildren} />
      {suspenseRemoteResource(
        projects,
        () =>
          projects?.projects && (
            <DataGrid
              autoHeight
              pagination
              className="bg-white"
              columns={columns}
              density="compact"
              disableSelectionOnClick
              disableDensitySelector
              disableColumnSelector
              disableColumnMenu
              rows={projects.projects}
            />
          )
      )}
    </>
  );
}
