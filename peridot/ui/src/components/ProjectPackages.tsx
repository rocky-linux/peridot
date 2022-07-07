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
import Alert from '@mui/material/Alert';
import TextField from '@mui/material/TextField';
import SearchIcon from '@mui/icons-material/Search';
import {
  DataGrid,
  GridColDef,
  GridColumns,
  GridRowsProp,
  GridSelectionModel,
  GridRenderCellParams,
} from '@mui/x-data-grid';
import { PageWrapper } from 'dotui/PageWrapper';
import { Table, TableCol, TableRow } from 'dotui/Table';
import {
  V1ListPackagesResponse,
  V1AsyncTask,
  V1PackageType,
} from 'bazel-bin/peridot/proto/v1/client_typescript';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import { buildApi, importApi, packageApi } from 'peridot/ui/src/api';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { Link } from 'react-router-dom';
import { PeridotLink } from 'common/ui/PeridotLink';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import { RemoteErrors } from 'common/ui/types';
import Snackbar from '@mui/material/Snackbar';
import Toolbar from '@mui/material/Toolbar';
import Divider from '@mui/material/Divider';
import Button from '@mui/material/Button';
import { reqap } from 'common/ui/reqap';
import to from 'await-to-js';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import { Header } from 'common/mui/Header';
import InputAdornment from '@mui/material/InputAdornment';

const columns: GridColDef[] = [
  {
    field: 'id',
    headerName: 'Name',
    sortable: false,
    flex: 1,
    renderCell: (params) => (
      <PeridotLink to={`/packages/${params.row['name']}`}>
        {params.row['name']}
      </PeridotLink>
    ),
  },
  {
    field: 'tags',
    headerName: 'Tags',
    sortable: false,
    flex: 1,
    renderCell: (params: GridRenderCellParams<string>) => {
      return (
        <Stack direction="row" spacing={1}>
          {(params.row['type'] === V1PackageType.NormalFork ||
            params.row['type'] === V1PackageType.NormalSrc ||
            params.row['type'] === V1PackageType.NormalForkModuleComponent ||
            params.row['type'] === V1PackageType.NormalForkModule) && (
            <Chip size="small" label="Package" variant="outlined" />
          )}
          {(params.row['type'] === V1PackageType.ModuleFork ||
            params.row['type'] === V1PackageType.ModuleForkModuleComponent) && (
            <Chip
              size="small"
              label="Module"
              color="success"
              variant="outlined"
            />
          )}
          {(params.row['type'] === V1PackageType.ModuleForkComponent ||
            params.row['type'] === V1PackageType.ModuleForkModuleComponent) && (
            <Chip
              size="small"
              label="Part of module"
              color="primary"
              variant="outlined"
            />
          )}
        </Stack>
      );
    },
  },
  {
    field: 'lastImportAt',
    headerName: 'Last import',
    sortable: false,
    flex: 1,
    renderCell: (params: GridRenderCellParams<string | null>) => (
      <Chip
        size="small"
        variant="outlined"
        label={params.value ? new Date(params.value).toLocaleString() : 'Never'}
        color={params.value ? 'success' : 'error'}
      />
    ),
  },
  {
    field: 'lastBuildAt',
    headerName: 'Last build',
    sortable: false,
    flex: 1,
    renderCell: (params: GridRenderCellParams<string | null>) => (
      <Chip
        size="small"
        variant="outlined"
        label={params.value ? new Date(params.value).toLocaleString() : 'Never'}
        color={params.value ? 'success' : 'error'}
      />
    ),
  },
];

export default function () {
  const project = React.useContext(ProjectContext);

  const [submitting, setSubmitting] = React.useState(false);
  const [buildDisabled, setBuildDisabled] = React.useState(false);
  const [importDisabled, setImportDisabled] = React.useState(false);
  const [snackbarOpen, setSnackbarOpen] = React.useState(false);
  const [snackbarMessage, setSnackbarMessage] = React.useState('');
  const [snackbarLink, setSnackbarLink] = React.useState<string | null>(null);
  const [snackbarSeverity, setSnackbarSeverity] = React.useState<
    'success' | 'error'
  >('success');
  const [buildPackageIds, setBuildPackageIds] = React.useState<string[]>([]);
  const [nameFilter, setNameFilter] = React.useState<string | undefined>();
  const [searchTimeout, setSearchTimeout] =
    React.useState<NodeJS.Timeout | null>(null);
  const [packagesRes, setPackagesRes] = React.useState<
    V1ListPackagesResponse | RemoteErrors | undefined | null
  >(undefined);
  const [pageSize, setPageSize] = React.useState(100);
  const [page, setPage] = React.useState(0);

  fetchRemoteResource(
    () =>
      packageApi.listPackages({
        projectId: project?.id || '',
        filtersName: nameFilter,
        limit: pageSize,
        page: page,
      }),
    setPackagesRes,
    false,
    [nameFilter, pageSize, page]
  );

  const handleClose = () => {
    setSnackbarOpen(false);
  };

  const openSnackbar = (
    message: string,
    severity: 'success' | 'error',
    link: string | null
  ) => {
    setSnackbarMessage(message);
    setSnackbarSeverity(severity);
    setSnackbarLink(link);
    setSnackbarOpen(true);
  };

  const onSelectionModelChange = (sm: GridSelectionModel) => {
    setImportDisabled(false);
    setBuildDisabled(false);
    setBuildPackageIds(sm.map((x) => x.toString()));
  };

  const doNameSearch = (val: string | null) => {
    if (!val || val.trim().length === 0) {
      setNameFilter(undefined);
      return;
    }

    setNameFilter(val);
  };

  const nameSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e || !e.currentTarget) {
      return;
    }

    if (searchTimeout) {
      clearTimeout(searchTimeout);
    }

    const val = e.currentTarget.value;
    setSearchTimeout(
      setTimeout(() => {
        doNameSearch(val);
      }, 200)
    );
  };

  const triggerImport = async () => {
    setSubmitting(true);
    setImportDisabled(true);

    let err, res;

    if (buildPackageIds.length === 1) {
      [err, res] = await to(
        importApi.importPackage({
          projectId: project?.id || '',
          body: {
            packageId: buildPackageIds[0],
          },
        })
      );
      if (!err && res) {
        openSnackbar('Import triggered', 'success', `tasks/${res.taskId}`);
      }
    } else {
      [err, res] = await to(
        importApi.importPackageBatch({
          projectId: project?.id || '',
          body: {
            imports: buildPackageIds.map((x) => ({ packageId: x })),
          },
        })
      );
      if (!err && res) {
        openSnackbar(
          'Imports triggered',
          'success',
          `import_batches/${res.importBatchId}`
        );
      }
    }

    if (err || !res) {
      openSnackbar('Could not trigger import(s)', 'error', null);
      setSubmitting(false);
      return;
    }

    setSubmitting(false);
  };

  const triggerBuild = async () => {
    setSubmitting(true);
    setBuildDisabled(true);

    let err, res;

    if (buildPackageIds.length === 1) {
      [err, res] = await to(
        buildApi.submitBuild({
          projectId: project?.id || '',
          body: {
            packageId: buildPackageIds[0],
          },
        })
      );
      if (!err && res) {
        openSnackbar('Build triggered', 'success', `tasks/${res.taskId}`);
      }
    } else {
      [err, res] = await to(
        buildApi.submitBuildBatch({
          projectId: project?.id || '',
          body: {
            builds: buildPackageIds.map((x) => ({
              packageId: x,
            })),
          },
        })
      );
      if (!err && res) {
        openSnackbar(
          'Builds triggered',
          'success',
          `build_batches/${res.buildBatchId}`
        );
      }
    }

    if (err || !res) {
      openSnackbar('Could not trigger build(s)', 'error', null);
      setSubmitting(false);
      return;
    }

    setSubmitting(false);
  };

  return (
    <>
      <Header title="Packages" />
      <Divider />
      <Toolbar className="bg-gray-100 flex space-x-6 py-3">
        <TextField
          fullWidth
          className="h-full"
          label="Search"
          type="search"
          size="small"
          onChange={nameSearch}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
        />
        {window.state.email && (
          <>
            <Button
              variant="contained"
              disabled={
                buildPackageIds.length === 0 || submitting || importDisabled
              }
              onClick={triggerImport}
            >
              Import
            </Button>
            <Button
              variant="contained"
              disabled={
                buildPackageIds.length === 0 || submitting || buildDisabled
              }
              onClick={triggerBuild}
            >
              Build
            </Button>
          </>
        )}
      </Toolbar>
      <Snackbar
        open={snackbarOpen}
        onClose={handleClose}
        autoHideDuration={5000}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert sx={{ width: '100%' }} severity={snackbarSeverity}>
          <div className="flex space-x-2">
            <div>{snackbarMessage}</div>
            {snackbarLink && (
              <div>
                <PeridotLink to={`/${snackbarLink}`}>Details</PeridotLink>
              </div>
            )}
          </div>
        </Alert>
      </Snackbar>
      {suspenseRemoteResource(packagesRes, (res: V1ListPackagesResponse) => {
        return (
          res.packages && (
            <DataGrid
              autoHeight
              pagination
              checkboxSelection={!!window.state.email}
              disableSelectionOnClick
              disableDensitySelector
              disableColumnSelector
              disableColumnMenu
              className="bg-white"
              columns={columns}
              density="compact"
              rows={res.packages}
              rowsPerPageOptions={[10, 25, 50, 100]}
              rowCount={parseInt(res.total || '0')}
              paginationMode="server"
              pageSize={pageSize}
              page={page}
              onPageChange={(page) => setPage(page)}
              onPageSizeChange={(newPageSize) => setPageSize(newPageSize)}
              onSelectionModelChange={onSelectionModelChange}
              selectionModel={buildPackageIds}
            />
          )
        );
      })}
    </>
  );
}
