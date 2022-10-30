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
  DataGrid,
  GridColDef,
  GridColumns,
  GridRowsProp,
} from '@mui/x-data-grid';
import {
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  CircularProgress,
  TextField,
  Container,
  Typography,
  Divider,
} from '@mui/material';
import { Link } from 'react-router-dom';

import {
  V1Advisory,
  V1ListAdvisoriesResponse,
} from 'bazel-bin/apollo/proto/v1/client_typescript/models';
import { api } from '../api';
import { reqap } from 'common/ui/reqap';
import { severityToBadge, typeToBadge } from 'apollo/ui/src/enumToText';

export const Overview = () => {
  // When advisories is set to null that means an error has occurred
  // Undefined means loading
  const [advisories, setAdvisories] = React.useState<
    V1Advisory[] | undefined | null
  >();
  const [pageSize, setPageSize] = React.useState(25);
  const [page, setPage] = React.useState(0);
  const [total, setTotal] = React.useState(0);
  const [filterSynopsis, setFilterSynopsis] = React.useState<
    string | undefined
  >();
  const [filterCve, setFilterCve] = React.useState<string | undefined>();

  React.useEffect(() => {
      const timer = setTimeout(() => {
        (async () => {
          let err, res: void | V1ListAdvisoriesResponse | undefined;
          [err, res] = await reqap(() =>
            api.listAdvisories({
              page,
              limit: pageSize,
              filtersSynopsis: filterSynopsis,
              filtersCve: filterCve,
            })
          );
          if (err || !res) {
            setAdvisories(null);
            return;
          }

          if (res) {
            setAdvisories(res.advisories);
            setTotal(parseInt(res.total || '0'));
          }
        })().then();
      }, 500);

      return () => clearTimeout(timer);
  }, [pageSize, page, filterSynopsis, filterCve]);

  const columns: GridColDef[] = [
    {
      field: 'id',
      headerName: 'Advisory',
      width: 150,
      sortable: false,
      renderCell: (params) => (
        <Link
          className="no-underline text-peridot-primary visited:text-purple-500"
          to={`/${params.value}`}
        >
          {params.value}
        </Link>
      ),
    },
    {
      field: 'synopsis',
      headerName: 'Synopsis',
      width: 450,
      sortable: false,
    },
    {
      field: 'severity',
      headerName: 'Severity',
      width: 150,
      sortable: false,
      renderCell: (params) => severityToBadge(params.value, 'small'),
    },
    {
      field: 'products',
      headerName: 'Products',
      width: 450,
      sortable: false,
    },
    {
      field: 'publish_date',
      headerName: 'Publish date',
      width: 170,
      sortable: false,
    },
  ];

  return (
    <div className="w-full">
      {advisories === undefined && <CircularProgress />}
      {advisories === null && (
        <h2 className="text-lg text-red-800 font-bold">
          Oh no! Something has gone wrong!
        </h2>
      )}
      {advisories && (
        <>
          <Container
            maxWidth={false}
            className="flex items-center space-x-4 bg-white"
            style={{ paddingTop: '0.5rem', paddingBottom: '0.5rem' }}
          >
            <Typography variant="overline">Filters</Typography>
            <TextField
              label="Synopsis"
              variant="outlined"
              size="small"
              onChange={(e) => setFilterSynopsis(e.target.value)}
            />
            <TextField
              label="CVE"
              variant="outlined"
              size="small"
              onChange={(e) => setFilterCve(e.target.value)}
            />
          </Container>
          <Divider />
          <Container maxWidth={false} disableGutters>
            <DataGrid
              autoHeight
              pagination
              disableSelectionOnClick
              disableDensitySelector
              disableColumnSelector
              disableColumnMenu
              className="bg-white"
              style={{ borderRadius: 0, border: 0 }}
              rows={advisories.map((advisory: V1Advisory) => ({
                id: advisory.name,
                synopsis: advisory.synopsis,
                severity: advisory.severity,
                products: advisory.affectedProducts?.join(', '),
                publish_date: Intl.DateTimeFormat('en-US', {
                  day: '2-digit',
                  month: 'short',
                  year: 'numeric',
                }).format(advisory.publishedAt),
              }))}
              rowsPerPageOptions={[10, 25, 50, 100]}
              rowCount={total}
              paginationMode="server"
              columns={columns}
              density="compact"
              pageSize={pageSize}
              onPageChange={(page) => setPage(page)}
              onPageSizeChange={(newPageSize) => setPageSize(newPageSize)}
            />
          </Container>
        </>
      )}
    </div>
  );
};
