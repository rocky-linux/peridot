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
import { V1SessionStatusResponse } from 'bazel-bin/obsidian/proto/v1/client_typescript';
import to from 'await-to-js';
import { api } from 'obsidian/ui/src/api';
import { RippleLoading } from 'dotui/RippleLoading';
import Button from '@mui/material/Button';

interface ScopeDescription {
  [key: string]: string;
}

const scopeToDescription: ScopeDescription = {
  profile: 'Read access to your basic profile information',
  email: 'Read access to your email address',
};

export default function () {
  const urlParams = new URLSearchParams(window.location.search);
  const consentChallenge = urlParams.get('consent_challenge');
  const [challenge, setChallenge] = React.useState<string | null>(
    consentChallenge
  );
  const [consentVerified, setConsentVerified] = React.useState(false);
  const [sessionStatus, setSessionStatus] =
    React.useState<V1SessionStatusResponse>();

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
            checkType: 'consent',
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

      setConsentVerified(true);
      setSessionStatus(res);
    }
  };

  const consentDecision = async (allow: boolean) => {
    let err, res: V1SessionStatusResponse | undefined;
    [err, res] = await to(
      api.consentDecision({
        body: {
          allow,
          challenge,
        },
      })
    );
    if (err || !res) {
      setChallenge(null);
      return;
    }

    window.location.href = res.redirectUrl;
  };

  return (
    <>
      {!consentVerified && !challenge && (
        <div>Invalid consent challenge, please try again</div>
      )}
      {!consentVerified && challenge && <RippleLoading />}
      {consentVerified && challenge && sessionStatus && (
        <div className="pt-12">
          <div className="font-bold text-xl pb-12 text-center">
            <span className="text-peridot-primary">
              {sessionStatus.clientName}
            </span>{' '}
            is requesting access to data
          </div>
          <div className="mx-auto">
            <div className="font-medium text-lg mb-4">
              The following is requested:
            </div>
            <div className="space-y-2">
              {sessionStatus.scopes?.map((scope: string) =>
                ['openid', 'offline_access'].includes(scope) ? null : (
                  <div className="flex items-center">
                    <div className="w-3 h-3 shadow rounded-full bg-peridot-primary mr-6" />
                    {scopeToDescription[scope.trim()] || 'Unknown scope'}
                  </div>
                )
              )}
            </div>
          </div>
          <div className="flex justify-between items-center pt-24">
            <Button onClick={() => consentDecision(false)}>Cancel</Button>
            <Button variant="contained" onClick={() => consentDecision(true)}>
              Allow
            </Button>
          </div>
        </div>
      )}
    </>
  );
}
