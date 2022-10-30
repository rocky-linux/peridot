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

import {
  AdvisorySeverity,
  V1AdvisoryType,
} from 'bazel-bin/apollo/proto/v1/client_typescript';
import Chip from '@mui/material/Chip';

export const severityToText = (severity?: AdvisorySeverity): string => {
  switch (severity) {
    case AdvisorySeverity.Critical:
      return 'Critical';
    case AdvisorySeverity.Important:
      return 'Important';
    case AdvisorySeverity.Moderate:
      return 'Moderate';
    case AdvisorySeverity.Low:
      return 'Low';
    default:
      return 'None';
  }
};

export const severityToBadge = (
  severity?: AdvisorySeverity,
  size?: 'small',
): React.ReactNode => {
  let color: 'primary' | 'secondary' | 'success' | 'info' | 'error' | 'warning' = 'success';

  switch (severity) {
    case AdvisorySeverity.Critical:
      color = 'error';
      break;
    case AdvisorySeverity.Important:
      color = 'warning';
      break;
    case AdvisorySeverity.Moderate:
      color = 'secondary';
      break;
    case AdvisorySeverity.Low:
      color = 'primary';
      break;
  }

  return <Chip label={severityToText(severity)} color={color} size={size} variant={size ? 'outlined' : undefined} />;
};

export const typeToText = (type?: V1AdvisoryType): string => {
  switch (type) {
    case V1AdvisoryType.Bugfix:
      return 'Bug Fix';
    case V1AdvisoryType.Security:
      return 'Security';
    case V1AdvisoryType.Enhancement:
      return 'Enhancement';
    default:
      return 'Unknown';
  }
};

export const typeToBadge = (
  type?: V1AdvisoryType,
  size?: 'small',
): React.ReactNode => {
  let color: 'info' | 'warning' = 'info';

  switch (type) {
    case V1AdvisoryType.Security:
      color = 'warning';
  }

  return <Chip label={typeToText(type)} color={color} size={size} variant={size ? 'outlined' : undefined} />;
};
