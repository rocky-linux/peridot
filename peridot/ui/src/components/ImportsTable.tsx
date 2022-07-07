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
import { DataGrid, GridColDef, GridRenderCellParams } from '@mui/x-data-grid';
import {
  V1Import,
  V1ImportRevision,
} from 'bazel-bin/peridot/proto/v1/client_typescript';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { transformTaskStatusToIcon } from 'peridot/ui/src/components/ProjectTasks';
import { PeridotLink } from 'common/ui/PeridotLink';

export interface ImportsTableProps {
  imports: V1Import[];
  total: number;
  page: number;
  pageSize: number;

  setPage?(page: number): void;

  setPageSize?(pageSize: number): void;
}

export const ImportsTable = (props: ImportsTableProps) => {
  const project = React.useContext(ProjectContext);

  const columns: GridColDef[] = [
    {
      field: 'status',
      headerName: 'Status',
      sortable: false,
      renderCell: (params) => transformTaskStatusToIcon(params.value),
    },
    {
      field: 'createdAt',
      headerName: 'Initiated',
      sortable: false,
      width: 250,
      renderCell: (params: GridRenderCellParams<Date>) => (
        <div>{params.value.toLocaleString()}</div>
      ),
    },
    {
      field: 'taskId',
      headerName: 'Task',
      sortable: false,
      width: 300,
      renderCell: (params) => (
        <PeridotLink to={`/tasks/${params.value}`}>{params.value}</PeridotLink>
      ),
    },
    {
      field: 'name',
      headerName: 'Package',
      sortable: false,
      width: 200,
      renderCell: (params) => (
        <PeridotLink to={`/packages/${params.value}`}>
          {params.value}
        </PeridotLink>
      ),
    },
    {
      field: 'revisions',
      headerName: 'Revisions',
      sortable: false,
      flex: 1,
      renderCell: (params) => {
        let revisionMapping: any = {};
        for (const revision of params.value) {
          if (!revisionMapping[revision.scmBranchName]) {
            revisionMapping[revision.scmBranchName] = [];
          }
          revisionMapping[revision.scmBranchName].push(revision);
        }

        return (
          <div className="flex space-x-2">
            {Object.keys(revisionMapping).map((branchName: string) => (
              <>
                <div className="flex">
                  {branchName} (
                  <div className="flex space-x-1">
                    {revisionMapping[branchName].map(
                      (revision: V1ImportRevision) => (
                        <PeridotLink
                          real
                          target="_blank"
                          to={`${project?.targetGitlabHost}/${
                            project?.targetPrefix
                          }/${revision.module ? 'modules' : 'rpms'}/${
                            params.row['name']
                          }/-/commit/${revision.scmHash}`}
                        >
                          {revision.module ? 'modules' : 'rpms'}
                        </PeridotLink>
                      )
                    )}
                  </div>
                  )
                </div>
              </>
            ))}
          </div>
        );
      },
    },
  ];

  return (
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
      rows={props.imports}
      rowsPerPageOptions={[10, 25, 50, 100]}
      rowCount={props.total}
      paginationMode="server"
      pageSize={props.pageSize}
      page={props.page}
      onPageChange={(page) => props.setPage && props.setPage(page)}
      onPageSizeChange={(newPageSize) =>
        props.setPageSize && props.setPageSize(newPageSize)
      }
    />
  );
};
