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
  endpointHttp,
  NS,
  svcNameHttp,
} from '../../../common/frontend_server/upstream.mjs';

export default async function run(webpackConfig) {
  const devFrontendUrl = 'https://id-dev.internal.pdev.resf.localhost';
  const envPublicUrl = process.env['OBSIDIAN_FRONTEND_HTTP_PUBLIC_URL'];
  const frontendUrl = process.env['RESF_NS'] ? envPublicUrl : devFrontendUrl;

  server({
    disableAuth: true,
    baseURL: frontendUrl,
    apis: {
      '/api': {
        prodApiUrl: endpointHttp(svcNameHttp('obsidian'), NS('obsidian')),
        devApiUrl: `https://id-api-dev.internal.pdev.resf.localhost`,
      },
    },
    port: 16000,
    webpackConfig,
  }).then();
}

if (process.env.NODE_ENV === 'production') {
  run().then();
}
