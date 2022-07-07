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
import { V1Project } from 'bazel-bin/peridot/proto/v1/client_typescript';
import { reqap } from 'common/ui/reqap';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';

const arches = ['i686', 'x86_64', 'aarch64', 'armhfp', 'ppc64le', 's390x'];

export interface ProjectFormProps {
  update?: boolean;
  project?: V1Project | null;
}

export const ProjectForm = (props: ProjectFormProps) => {
  const [formArches, setFormArches] = React.useState<string[]>(
    props.project?.archs ?? []
  );
  const [targetVendor, setTargetVendor] = React.useState<string>(
    props.project?.targetVendor ?? 'redhat'
  );
  const [followImportDist, setFollowImportDist] = React.useState(
    props.project?.followImportDist ?? false
  );
  const [gitMakePublic, setGitMakePublic] = React.useState(
    props.project?.gitMakePublic ?? false
  );
  const [submitting, setSubmitting] = React.useState(false);
  const [errorMessage, setErrorMessage] = React.useState<string | null>(null);
  const [successMessage, setSuccessMessage] = React.useState<string | null>(
    null
  );

  const handleArchChange = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const {
      target: { value },
    } = event;
    setFormArches(
      // On autofill we get a stringified value.
      typeof value === 'string' ? value.split(',') : value
    );
  };

  const handleVendorChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setTargetVendor(event.target.value);
  };

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
        }, {}),
      {
        archs: formArches,
        followImportDist,
        gitMakePublic,
        targetVendor,
      }
    );

    if (props.update) {
      const [err] = await reqap(() =>
        projectApi.updateProject({
          projectId: props.project?.id ?? '',
          body: {
            project: body,
          },
        })
      );
      if (err) {
        setErrorMessage((await err).message);
        setSubmitting(false);
        return;
      }

      setSuccessMessage('Project updated successfully');
    } else {
      const [err, res] = await to(
        projectApi.createProject({
          body: {
            project: body,
          },
        })
      );
      if (err) {
        setErrorMessage(err.message);
        setSubmitting(false);
        return;
      }

      window.location.href = '/' + res?.project?.id;
    }
    setSubmitting(false);
  };

  return (
    <>
      {errorMessage && <Alert severity="error">{errorMessage}</Alert>}
      {successMessage && <Alert severity="success">{successMessage}</Alert>}
      <form onSubmit={onSubmit}>
        <div className="flex">
          <div className="space-y-4 w-1/2 pr-12">
            <p className="font-bold">Details</p>
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Name"
              name="name"
              defaultValue={props.project?.name}
            />
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              type="number"
              label="Major version"
              InputProps={{ inputProps: { min: 0 } }}
              name="majorVersion"
              defaultValue={props.project?.majorVersion}
            />
            <TextField
              select
              SelectProps={{
                multiple: true,
              }}
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Architectures"
              id="archs"
              name="archs"
              value={formArches}
              onChange={handleArchChange}
            >
              {arches.map((arch) => (
                <MenuItem key={arch} value={arch}>
                  {arch}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Dist tag"
              name="distTag"
              defaultValue={props.project?.distTag}
            />
            <TextField
              select
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Target vendor"
              id="targetVendor"
              name="targetVendor"
              defaultValue={props.project?.targetVendor}
              value={targetVendor}
              onChange={handleVendorChange}
            >
              <MenuItem value="redhat">Red Hat</MenuItem>
              <MenuItem value="suse">SUSE</MenuItem>
            </TextField>
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Additional vendor"
              name="additionalVendor"
              defaultValue={props.project?.additionalVendor}
            />
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Vendor macro"
              name="vendorMacro"
              defaultValue={props.project?.vendorMacro}
            />
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Packager macro"
              name="packagerMacro"
              defaultValue={props.project?.packagerMacro}
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={followImportDist}
                  onChange={(event) =>
                    setFollowImportDist(event.target.checked)
                  }
                  name="followImportDist"
                />
              }
              label="Follow import dist tag"
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={gitMakePublic}
                  onChange={(event) =>
                    setGitMakePublic(event.target.checked)
                  }
                  name="gitMakePublic"
                />
              }
              label="Make import repositories public"
            />
          </div>
          <div className="space-y-4 w-1/2">
            <p className="font-bold">Source control</p>
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Target Gitlab host"
              name="targetGitlabHost"
              defaultValue={props.project?.targetGitlabHost}
            />
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Target prefix"
              name="targetPrefix"
              defaultValue={props.project?.targetPrefix}
            />
            <TextField
              sx={{ display: 'flex' }}
              required
              size="small"
              label="Target branch prefix"
              name="targetBranchPrefix"
              defaultValue={props.project?.targetBranchPrefix}
            />
            <TextField
              sx={{ display: 'flex' }}
              size="small"
              label="Source git host"
              name="sourceGitHost"
              defaultValue={props.project?.sourceGitHost}
            />
            <TextField
              sx={{ display: 'flex' }}
              size="small"
              label="Source prefix"
              name="sourcePrefix"
              defaultValue={props.project?.sourcePrefix}
            />
            <TextField
              sx={{ display: 'flex' }}
              size="small"
              label="Source branch prefix"
              name="sourceBranchPrefix"
              defaultValue={props.project?.sourceBranchPrefix}
            />
            <TextField
              sx={{ display: 'flex' }}
              size="small"
              label="Branch suffix"
              name="branchSuffix"
              defaultValue={props.project?.branchSuffix}
            />
          </div>
        </div>
        <div className="w-full flex justify-end mt-4">
          <Button type="submit" disabled={submitting} variant="contained">
            {props.update ? 'Update' : 'Create'}
          </Button>
        </div>
      </form>
    </>
  );
};
