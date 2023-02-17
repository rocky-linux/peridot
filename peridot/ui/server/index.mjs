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

// noinspection ES6PreferShortImport

import server from '../../../common/frontend_server/index.mjs';
import {
  svcNameHttp,
  endpointHttp,
  NS,
} from '../../../common/frontend_server/upstream.mjs';
import {
  hydraAutoSignup,
  hydraPublic,
} from '../../../hydra/pkg/hydra/autosignup.mjs';

export default async function run(webpackConfig) {
  const devFrontendUrl = 'http://peridot.pdot.localhost:15000';
  const envPublicUrl = process.env['PERIDOT_FRONTEND_HTTP_PUBLIC_URL'];
  const frontendUrl = process.env['RESF_NS'] ? envPublicUrl : devFrontendUrl;

  const wellKnown = await hydraPublic.discoverOidcConfiguration();
  const hdr = await hydraAutoSignup({
    name: 'Peridot',
    client: 'peridot',
    internal: false,
    frontend: true,
    scopes: 'openid profile email offline_access',
    redirectUri: `${frontendUrl}/oauth2/callback`,
    postLogoutRedirectUri: frontendUrl,
  });

  server({
    issuerBaseURL: wellKnown.data.issuer,
    clientID: hdr.clientID,
    clientSecret: hdr.secret,
    baseURL: frontendUrl,
    apis: {
      '/api': {
        prodApiUrl: endpointHttp(
          svcNameHttp('peridotserver'),
          NS('peridotserver')
        ),
        devApiUrl: `https://peridot-api-dev.internal.pdev.resf.localhost`,
      },
    },
    port: 15000,
    disableAuthEnforce: true,
    webpackConfig,
  }).then();
}

if (process.env.NODE_ENV === 'production') {
  run().then();
}
