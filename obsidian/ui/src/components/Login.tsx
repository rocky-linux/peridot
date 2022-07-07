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
import { GoogleLoginButton } from 'react-social-login-buttons';
import {
  V1SessionStatusResponse,
  V1GetOAuth2ProvidersResponse,
  V1OAuth2Provider,
} from 'bazel-bin/obsidian/proto/v1/client_typescript';
import to from 'await-to-js';
import { api } from 'obsidian/ui/src/api';
import { RippleLoading } from 'dotui/RippleLoading';

export default function () {
  const urlParams = new URLSearchParams(window.location.search);
  const loginChallenge = urlParams.get('login_challenge');
  const [challenge, setChallenge] = React.useState<string | null>(
    loginChallenge
  );
  const [loginVerified, setLoginVerified] = React.useState(false);
  const [providers, setProviders] = React.useState<V1OAuth2Provider[] | null>();

  React.useEffect(() => {
    checkQueryString().then();
  }, []);

  const checkQueryString = async () => {
    if (challenge && challenge !== '') {
      let err, res: V1SessionStatusResponse | undefined;
      [err, res] = await to(
        api.sessionStatus({
          body: {
            challenge,
            checkType: 'login',
          },
        })
      );
      if (err || !res || !res.valid) {
        setChallenge(null);
        return;
      }

      if (res.redirectUrl) {
        window.location.href = res.redirectUrl;
        return;
      }

      let err2, res2: V1GetOAuth2ProvidersResponse | undefined;
      [err2, res2] = await to(api.getOAuth2Providers());
      if (err2 || !res2) {
        setChallenge(null);
        return;
      }

      setProviders(res2.providers);

      setLoginVerified(true);
    }
  };

  return (
    <>
      {!loginVerified && !challenge && (
        <div>Invalid login challenge, please try again</div>
      )}
      {!loginVerified && challenge && <RippleLoading />}
      {loginVerified && challenge && (
        <>
          <div className="font-medium text-lg mb-4 text-center">
            Choose your login provider
          </div>
          {providers &&
            providers.map((provider) => {
              const redirectToInitiation = () => {
                window.location.href = `/api/v1/oauth2/initiate_session?provider_id=${provider.id}&challenge=${challenge}`;
              };

              switch (provider.provider) {
                case 'google':
                  return <GoogleLoginButton onClick={redirectToInitiation} />;
                default:
                  return 'Invalid provider';
              }
            })}
        </>
      )}
    </>
  );
}
