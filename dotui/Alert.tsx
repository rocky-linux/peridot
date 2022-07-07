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
import classnames from 'classnames';
import { Intent } from './intent';

export interface AlertProps {
  children?: React.ReactNode;
  intent?: Intent;
  className?: string;
}

export const Alert = (props: AlertProps) => {
  const finalIntent = props.intent || Intent.NEUTRAL;

  let intentClass = 'border-gray-300 bg-white';
  switch (finalIntent) {
    case Intent.PRIMARY:
      intentClass = 'border-peridot-primary bg-peridot-primary bg-opacity-30';
      break;
    case Intent.ERROR:
      intentClass = 'border-red-300 bg-red-300 bg-opacity-30';
      break;
    case Intent.WARNING:
      intentClass = 'border-yellow-300 bg-yellow-300 bg-opacity-30';
      break;
    case Intent.SUCCESS:
      intentClass = 'border-green-300 bg-green-300 bg-opacity-30';
      break;
  }

  return (
    <div
      className={classnames('p-2 rounded border', intentClass, props.className)}
    >
      {props.children}
    </div>
  );
};
