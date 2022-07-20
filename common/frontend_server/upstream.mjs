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

import os from 'os';

export function envOverridable(svcName, protocol, x) {
  const envName = `${svcName}_${protocol}_ENDPOINT_OVERRIDE`.toUpperCase();
  const envValue = process.env[envName];
  if (envValue) {
    return envValue;
  }
  return x();
}

export function svcName(svc, protocol) {
  let env = process.env['BYC_ENV'];
  if (!env) {
    env = 'dev';
  }
  return `${svc}-${protocol}-${env}-service`;
}

export function svcNameHttp(svc) {
  return svcName(svc, 'http');
}

export function endpoint(generatedServiceName, ns, port) {
  const forceNs = process.env['BYC_FORCE_NS'];
  if (forceNs) {
    ns = forceNs;
  }
  return `${generatedServiceName}.${ns}.svc.cluster.local${port}`;
}

export function endpointHttp(generatedServiceName, ns, port = '') {
  // noinspection HttpUrlsUsage
  return `http://${endpoint(generatedServiceName, ns, port)}`;
}

export function NS(ns) {
  const env = process.env['BYC_ENV'];
  if (!env || env === 'dev') {
    const bycNs = process.env['BYC_NS'];
    if (!bycNs) {
      return `${os.userInfo().username}-dev`;
    }
    return bycNs;
  }

  return ns;
}
