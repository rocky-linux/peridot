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
import classnames from 'classnames';
import {
  DataGrid,
  GridColDef,
  GridColumns,
  GridRowsProp,
} from '@mui/x-data-grid';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ChangeCircleIcon from '@mui/icons-material/ChangeCircle';
import PendingIcon from '@mui/icons-material/Pending';
import CancelIcon from '@mui/icons-material/Cancel';
import { PageWrapper } from 'dotui/PageWrapper';
import { Table, TableCol, TableRow } from 'dotui/Table';
import {
  V1ListPackagesResponse,
  V1Package,
  V1ListTasksResponse,
  V1TaskStatus,
  V1TaskType,
  ApiResponse,
  V1AsyncTask,
} from 'bazel-bin/peridot/proto/v1/client_typescript';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import { packageApi, taskApi } from 'peridot/ui/src/api';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { Link } from 'react-router-dom';
import { TextField } from 'dotui/TextField';
import { PeridotLink } from 'common/ui/PeridotLink';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import Chip from '@mui/material/Chip';
import { Header } from 'common/mui/Header';
import { RemoteErrors } from 'common/ui/types';

export const transformTaskType = (val: string) => {
  switch (val) {
    case V1TaskType.Build:
      return (
        <Chip size="small" variant="outlined" label="Build" color="success" />
      );
    case V1TaskType.Import:
      return (
        <Chip size="small" variant="outlined" label="Import" color="primary" />
      );
    default:
      return val;
  }
};

export const translateTaskTypeToText = (type: V1TaskType) => {
  switch (type) {
    case 'TASK_TYPE_UNSPECIFIED':
      return 'Unspecified';
    case 'TASK_TYPE_IMPORT':
      return 'Import';
    case 'TASK_TYPE_IMPORT_SRC_GIT':
      return 'Package src-git tarballs';
    case 'TASK_TYPE_IMPORT_SRC_GIT_TO_DIST_GIT':
      return 'Push src-git to dist-git';
    case 'TASK_TYPE_IMPORT_UPSTREAM':
      return 'Import from upstream';
    case 'TASK_TYPE_BUILD':
      return 'Build';
    case 'TASK_TYPE_BUILD_SRPM':
      return 'Build (SRPM)';
    case 'TASK_TYPE_BUILD_ARCH':
      return 'Build (Arch)';
    case 'TASK_TYPE_BUILD_SRPM_UPLOAD':
      return 'Upload SRPM';
    case 'TASK_TYPE_BUILD_ARCH_UPLOAD':
      return 'Upload Arch';
    case 'TASK_TYPE_WORKER_PROVISION':
      return 'Provision ephemeral worker';
    case 'TASK_TYPE_WORKER_DESTROY':
      return 'Destroy ephemeral worker';
    case 'TASK_TYPE_YUMREPOFS_UPDATE':
      return 'Push updates to Yumrepofs';
    case 'TASK_TYPE_KEYKEEPER_SIGN_ARTIFACT':
      return 'Sign artifact';
    case 'TASK_TYPE_SYNC_CATALOG':
      return 'Catalog sync';
    default:
      return 'Unknown';
  }
};

export const transformTaskStatusToIcon = (
  val:
    | V1TaskStatus
    | string
    | number
    | false
    | true
    | object
    | undefined
    | null,
  includeText?: boolean,
  prefix?: string,
) => {
  let icon = <></>;
  let baseClasses = 'flex items-center space-x-1';
  let statusText = 'Pending';

  switch (val) {
    case V1TaskStatus.Canceled:
      baseClasses = classnames(baseClasses, 'text-yellow-300');
      icon = <CancelIcon />;
      statusText = 'Canceled';
      break;
    case V1TaskStatus.Failed:
      baseClasses = classnames(baseClasses, 'text-red-500');
      icon = <CancelIcon />;
      statusText = 'Failed';
      break;
    case V1TaskStatus.Running:
      baseClasses = classnames(baseClasses, 'text-peridot-primary');
      icon = <ChangeCircleIcon />;
      statusText = 'Running';
      break;
    case V1TaskStatus.Succeeded:
      baseClasses = classnames(baseClasses, 'text-green-500');
      icon = <CheckCircleIcon />;
      statusText = 'Succeeded';
      break
    default:
      baseClasses = classnames(baseClasses, 'text-yellow-600');
      icon = <PendingIcon />;
      statusText = 'Pending';
      break;
  }

  return (
    <div className={baseClasses}>
      {icon}
      {includeText && <div className="text-black text-xl">{prefix} {statusText}</div>}
    </div>
  );
};

const columns: GridColDef[] = [
  {
    field: 'status',
    headerName: 'Status',
    sortable: false,
    renderCell: (params) => transformTaskStatusToIcon(params.value),
  },
  {
    field: 'id',
    headerName: 'ID',
    sortable: false,
    flex: 1,
    renderCell: (params) => (
      <PeridotLink to={`/tasks/${params.value}`}>{params.value}</PeridotLink>
    ),
  },
  {
    field: 'type',
    headerName: 'Type',
    sortable: false,
    flex: 1,
    renderCell: (params) => transformTaskType(params.value?.toString() || ''),
  },
  {
    field: 'metadata',
    headerName: 'Metadata',
    sortable: false,
    flex: 1,
    renderCell: (params) => {
      const val: any = params.value;
      return <code>packageName={val?.packageName}</code>;
    },
  },
  {
    field: 'submitter',
    headerName: 'Submitter',
    sortable: false,
    flex: 1,
    renderCell: (params) => {
      return params.row["submitterDisplayName"] || params.row["submitterEmail"] || params.row["submitterId"];
    },
  },
];

export default function () {
  const project = React.useContext(ProjectContext);

  const [tasksRes, setTasksRes] = React.useState<
    V1ListTasksResponse | RemoteErrors | undefined | null
  >(undefined);
  const [pageSize, setPageSize] = React.useState(100);
  const [page, setPage] = React.useState(0);

  fetchRemoteResource(
    () =>
      taskApi.listTasks({
        projectId: project?.id || '',
        limit: pageSize,
        page,
      }),
    setTasksRes,
    false,
    [pageSize, page]
  );

  return (
    <>
      <Header title="Tasks" />
      {suspenseRemoteResource(tasksRes, (res: V1ListTasksResponse) => {
        return (
          res.tasks && (
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
              rows={res.tasks.map((task: V1AsyncTask) =>
                task.subtasks ? task.subtasks[0] : { id: 'invalid' }
              )}
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
      })}
    </>
  );
}
