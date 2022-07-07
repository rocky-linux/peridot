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

import Alert from '@mui/material/Alert';
import { H1 } from 'dotui/H1';
import { H2 } from 'dotui/H2';
import { H4 } from 'dotui/H4';
import { H5 } from 'dotui/H5';
import { PageWrapper } from 'dotui/PageWrapper';
import React from 'react';
import { reqap } from './reqap';
import { RemoteState, AccessDeniedError } from './types';

export var loadingElement: React.ReactElement = <div>Loading...</div>;

export const setLoadingElement = (elem: React.ReactElement) => {
  loadingElement = elem;
};

export function fetchRemoteResource<T, X>(
  fetchFunction: () => Promise<T>,
  setFunction: (val: RemoteState<T>) => void,
  disableEffect: boolean = false,
  effectArray: any[] = []
) {
  const remoteMethod = async () => {
    const [err, res] = await reqap(() => fetchFunction());
    if (err) {
      if (err.status === 403) {
        setFunction('access_denied');
      } else {
        setFunction(null);
      }
    }

    if (res) {
      setFunction(res);
    }
  };

  if (disableEffect) {
    remoteMethod().then();
  } else {
    React.useEffect(() => {
      remoteMethod().then();
    }, effectArray);
  }
}

export function suspenseRemoteResource<T>(
  state: RemoteState<T>,
  render: (res: T) => React.ReactNode,
  wrap: ((elem: React.ReactNode) => React.ReactNode) | undefined = undefined
): React.ReactNode {
  const wrapHelper = (elem: React.ReactNode) => {
    if (wrap) {
      return wrap(elem);
    }

    return elem;
  };

  if (state === undefined) {
    return wrapHelper(loadingElement);
  } else if (state === 'access_denied') {
    return wrapHelper(
      <Alert severity="error">
        Access to resource has been denied. Contact an administrator if you
        think this is a mistake
      </Alert>
    );
  } else if (state === null) {
    return wrapHelper(
      <div className="flex items-center justify-center flex-col mt-4">
        <H2>Could not fetch remote resource</H2>
        <H4>Please try again later</H4>
      </div>
    );
  } else {
    return render(state);
  }
}
