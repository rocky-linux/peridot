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
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import { PageWrapper } from 'dotui/PageWrapper';
import Typography from '@mui/material/Typography';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Button from '@mui/material/Button';
import to from 'await-to-js';
import { projectApi } from 'peridot/ui/src/api';
import Alert from '@mui/material/Alert';
import { reqap } from 'common/ui/reqap';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { useHistory } from 'react-router';

export interface SyncFormProps {}

export const SyncForm = (props: SyncFormProps) => {
  const history = useHistory();

  const project = React.useContext(ProjectContext);
  const [submitting, setSubmitting] = React.useState(false);
  const [errorMessage, setErrorMessage] = React.useState<string | null>(null);
  const [successMessage, setSuccessMessage] = React.useState<string | null>(
    null
  );

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    setSubmitting(true);
    setErrorMessage(null);

    event.preventDefault();
    const targets: HTMLInputElement[] = Array.from(event.target as any);
    const body = Object.assign(
      {},
      targets
        .map((el) => {
          return { name: el.name, value: el.value };
        })
        .reduce((acc: any, cur) => {
          if (cur.name && cur.name !== '') {
            acc[cur.name] = cur.value;
          }
          return acc;
        }, {})
    );

    const [err, res] = await to(
      projectApi.syncCatalog({
        projectId: project?.id ?? '',
        body,
      })
    );
    if (err) {
      setErrorMessage(err.message);
      setSubmitting(false);
      return;
    }

    history.push('/tasks/' + res?.taskId);
    setSubmitting(false);
  };

  return (
    <>
      {errorMessage && <Alert severity="error">{errorMessage}</Alert>}
      {successMessage && <Alert severity="success">{successMessage}</Alert>}
      <form onSubmit={onSubmit}>
        <div className="flex">
          <div className="space-y-4 w-full pr-12">
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="SCM URL"
              name="scmUrl"
            />
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Branch"
              name="branch"
            />
          </div>
        </div>
        <div className="w-full flex justify-end mt-4">
          <Button type="submit" disabled={submitting} variant="contained">
            Sync
          </Button>
        </div>
      </form>
    </>
  );
};
