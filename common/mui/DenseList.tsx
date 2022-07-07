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
import List from '@mui/material/List';
import ListItemText from '@mui/material/ListItemText';

export interface DenseListEntry {
  key?: string;
  value?: string;
  entries?: DenseListEntry[];
}

export interface DenseListProps {
  entries: DenseListEntry[];
}

const listEntry = (
  nestLevel: number,
  key?: string,
  value?: string,
  entries?: DenseListEntry[]
): React.ReactNode => {
  return (
    <>
      <div className="flex justify-between items-center">
        {key && (
          <ListItemText
            classes={{ primary: 'text-sm font-bold' }}
            sx={{ pl: 6 * nestLevel }}
            primary={key}
          />
        )}
        {value && (
          <ListItemText
            classes={{ secondary: 'text-sm' }}
            sx={key ? undefined : { pl: 6 * nestLevel }}
            secondary={value}
          />
        )}
      </div>
      {entries?.map((entry) =>
        listEntry(nestLevel + 1, entry.key, entry.value, entry.entries)
      )}
    </>
  );
};

export const DenseList = (props: DenseListProps) => {
  return (
    <List dense className="w-96 divide-y">
      {props.entries.map((entry) =>
        listEntry(0, entry.key, entry.value, entry.entries)
      )}
    </List>
  );
};
