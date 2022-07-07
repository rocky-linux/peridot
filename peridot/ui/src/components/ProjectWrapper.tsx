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

import { generateSuspensedComponent } from 'common/ui/SuspensedComponent';
import Alert from '@mui/material/Alert';
import { RippleLoading } from 'dotui/RippleLoading';
import React from 'react';
import { Route, Switch } from 'react-router-dom';
import { ProjectContext } from '../context/ProjectContext';
import NotFound from 'peridot/ui/src/components/NotFound';

export default function () {
  const project = React.useContext(ProjectContext);

  if (!project) {
    return (
      <>
        {project === undefined && <RippleLoading />}
        {project === null && (
          <Alert severity="error">Project could not be found</Alert>
        )}
      </>
    );
  }

  return (
    <>
      <Switch>
        <Route
          path="/packages/:name"
          component={generateSuspensedComponent(
            () => import('./ProjectPackageDetails')
          )}
        />
        <Route
          path="/packages"
          exact
          component={generateSuspensedComponent(
            () => import('./ProjectPackages')
          )}
        />
        <Route
          path="/tasks/:id"
          component={generateSuspensedComponent(
            () => import('./ProjectTaskDetails')
          )}
        />
        <Route
          path="/tasks"
          component={generateSuspensedComponent(() => import('./ProjectTasks'))}
        />
        <Route
          path="/imports"
          component={generateSuspensedComponent(
            () => import('./ProjectImports')
          )}
        />
        <Route
          path="/builds"
          component={generateSuspensedComponent(
            () => import('./ProjectBuilds')
          )}
        />
        <Route
          path="/build_batches/:id"
          component={generateSuspensedComponent(
            () => import('./ProjectBuildBatchShow')
          )}
        />
        <Route
          path="/build_batches"
          component={generateSuspensedComponent(
            () => import('./ProjectBuildBatches')
          )}
        />
        <Route
          path="/import_batches/:id"
          component={generateSuspensedComponent(
            () => import('./ProjectImportBatchShow')
          )}
        />
        <Route
          path="/import_batches"
          component={generateSuspensedComponent(
            () => import('./ProjectImportBatches')
          )}
        />
        <Route
          path="/settings"
          component={generateSuspensedComponent(
            () => import('./ProjectSettings')
          )}
        />
        <Route
          path="/credentials"
          component={generateSuspensedComponent(
            () => import('./ProjectCredentials')
          )}
        />
        <Route
          path="/repositories/:id"
          component={generateSuspensedComponent(
            () => import('./ProjectRepositoryShow')
          )}
        />
        <Route
          path="/repositories"
          component={generateSuspensedComponent(
            () => import('./ProjectRepositories')
          )}
        />
        <Route
          path="/sync"
          component={generateSuspensedComponent(
            () => import('./ProjectSync')
          )}
        />
        <Route component={NotFound} />
      </Switch>
    </>
  );
}
