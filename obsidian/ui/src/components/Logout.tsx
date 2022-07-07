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

export default function () {
  const urlParams = new URLSearchParams(window.location.search);
  const logoutChallenge = urlParams.get('logout_challenge');
  const [challenge, setChallenge] = React.useState<string | null>(
    logoutChallenge
  );

  const logoutDecision = async (accept: boolean) => {
    let err, res: V1SessionStatusResponse | undefined;
    [err, res] = await to(
      api.logoutDecision({
        body: {
          accept,
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
      {challenge && (
        <div className="pt-12">
          <div className="font-bold text-xl pb-12 text-center">
            A client requested to log you out
          </div>
          <div className="mx-auto">
            <div className="font-medium text-lg mb-4">
              The following is requested:
            </div>
            <div className="space-y-2">
              <div className="flex items-center">
                <div className="w-3 h-3 shadow rounded-full bg-peridot-primary mr-6" />
                End all active sessions for this user
              </div>
            </div>
          </div>
          <div className="flex justify-between items-center pt-24">
            <Button onClick={() => logoutDecision(false)}>Cancel</Button>
            <Button variant="contained" onClick={() => logoutDecision(true)}>
              Allow
            </Button>
          </div>
        </div>
      )}
    </>
  );
}
