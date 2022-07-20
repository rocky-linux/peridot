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

// noinspection JSUnresolvedFunction
// noinspection ES6PreferShortImport

import {
  svcNameHttp,
  endpointHttp,
  NS,
  envOverridable
} from '../../../common/frontend_server/upstream.mjs';
import pkg from '@ory/hydra-client';
import os from 'os';

const { Configuration, PublicApi, AdminApi } = pkg;

export function hydraPublicUrl() {
  return envOverridable('hydra_public', 'http', () => {
    const svc = svcNameHttp('hydra-public');
    return endpointHttp(svc, NS('hydra-public'), ':4444');
  });
}

function hydraAdminUrl() {
  return envOverridable('hydra_admin', 'http', () => {
    const svc = svcNameHttp('hydra-admin');
    return endpointHttp(svc, NS('hydra-admin'), ':4445');
  });
}

const hydraAdmin = new AdminApi(
  new Configuration({
    basePath: hydraAdminUrl()
  })
);

export const hydraPublic = new PublicApi(
  new Configuration({
    basePath: hydraPublicUrl()
  })
);

function secret() {
  const env = process.env['BYC_ENV'];
  if (!env || env === 'dev') {
    return 'dev-123-secret';
  }

  const scr = process.env['HYDRA_SECRET'];
  if (!scr || scr === '' || scr.length === 0) {
    throw 'HYDRA_SECRET is not set';
  }

  return scr;
}

export async function hydraAutoSignup(req) {
  let ns = process.env['BYC_NS'];
  if (!ns || ns === '') {
    ns = 'dev';
  }
  let name = `${req.client}-${ns}`;
  const serviceName = `autos-${name}`;
  if (req.name) {
    name = req.name;
  }
  const clientModel = {
    client_name: name,
    client_id: serviceName,
    scope: req.scopes,
    client_secret: secret(),
    redirect_uris: null,
    grant_types: ['authorization_code', 'refresh_token'],
  };
  if (req.frontend) {
    clientModel.redirect_uris = [req.redirectUri];
    clientModel.post_logout_redirect_uris = [req.postLogoutRedirectUri];
  }

  const ret = {
    clientID: serviceName,
    secret: secret()
  };

  try {
    await hydraAdmin.getOAuth2Client(serviceName);
    try {
      console.log(`Updated client ${name}`);
      await hydraAdmin.updateOAuth2Client(serviceName, clientModel);
    } catch (e) {
      // noinspection ExceptionCaughtLocallyJS
      throw e;
    }
  } catch (e) {
    console.log(`Created client ${name}`);
    await hydraAdmin.createOAuth2Client(clientModel);
  }

  return ret;
}
