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
import { RemoteErrors } from 'common/ui/types';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import { importApi, packageApi } from 'peridot/ui/src/api';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { V1ListImportBatchesResponse } from 'bazel-bin/peridot/proto/v1/client_typescript';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import { PeridotLink } from 'common/ui/PeridotLink';
import Divider from '@mui/material/Divider';
import { DataGrid, GridColDef, GridRenderCellParams } from '@mui/x-data-grid';
import { transformTaskStatusToIcon } from 'peridot/ui/src/components/ProjectTasks';
import { Header } from 'common/mui/Header';

export default function () {
  const project = React.useContext(ProjectContext);

  const columns: GridColDef[] = [
    {
      field: 'id',
      headerName: 'ID',
      sortable: false,
      width: 300,
      renderCell: (params) => (
        <PeridotLink to={`/import_batches/${params.value}`}>
          {params.value}
        </PeridotLink>
      ),
    },
    {
      field: 'count',
      headerName: 'Count',
      sortable: false,
    },
    {
      field: 'pending',
      headerName: 'Pending',
      sortable: false,
    },
    {
      field: 'running',
      headerName: 'Running',
      sortable: false,
    },
    {
      field: 'succeeded',
      headerName: 'Succeeded',
      sortable: false,
      width: 110,
    },
    {
      field: 'failed',
      headerName: 'Failed',
      sortable: false,
    },
    {
      field: 'canceled',
      headerName: 'Canceled',
      sortable: false,
    },
  ];

  const [importBatchesRes, setImportBatchesRes] = React.useState<
    V1ListImportBatchesResponse | RemoteErrors | undefined | null
  >(undefined);
  const [pageSize, setPageSize] = React.useState(100);
  const [page, setPage] = React.useState(0);

  fetchRemoteResource(
    () =>
      importApi.listImportBatches({
        projectId: project?.id || '',
        limit: pageSize,
        page: page.toString(),
      }),
    setImportBatchesRes,
    false,
    [pageSize, page]
  );

  return (
    <>
      <Header title="Import batches" />
      {suspenseRemoteResource(
        importBatchesRes,
        (res: V1ListImportBatchesResponse) => {
          return (
            res.importBatches && (
              <DataGrid
                autoHeight
                pagination
                disableSelectionOnClick
                disableDensitySelector
                disableColumnSelector
                disableColumnMenu
                className="bg-white"
                columns={columns}
                density="compact"
                rows={res.importBatches}
                rowsPerPageOptions={[10, 25, 50, 100]}
                rowCount={res.total}
                paginationMode="server"
                pageSize={pageSize}
                page={page}
                onPageChange={(page) => setPage(page)}
                onPageSizeChange={(newPageSize) => setPageSize(newPageSize)}
              />
            )
          );
        }
      )}
    </>
  );
}
