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

export interface TableCols {
  label: string;
  className?: string;
}

export interface TableProps {
  cols: TableCols[];
  children?: React.ReactNode;
}

export interface TableRowProps {
  children?: React.ReactNode;
  className?: string;
  hover?: boolean;
}

export interface TableColProps {
  children?: React.ReactNode;
  className?: string;
}

export const Table = (props: TableProps) => {
  return (
    <div className="flex flex-col">
      <div className="overflow-x-auto align-middle shadow sm:rounded">
        <table className="w-full">
          <thead>
            <TableRow className="bg-gray-50">
              {props.cols.map((col) => (
                <TableCol
                  className={classnames(
                    'font-medium text-gray-500 tracking-wide',
                    col.className
                  )}
                >
                  {col.label}
                </TableCol>
              ))}
            </TableRow>
          </thead>
          <tbody className="bg-white">{props.children}</tbody>
        </table>
      </div>
    </div>
  );
};

export const TableRow = (props: TableRowProps) => {
  return (
    <tr
      className={classnames(
        'divide-x',
        props.hover && 'hover:bg-gray-50 cursor-pointer',
        props.className
      )}
    >
      {props.children}
    </tr>
  );
};

export const TableCol = (props: TableColProps) => {
  return (
    <td
      className={classnames('p-3 text-sm whitespace-nowrap', props.className)}
    >
      {props.children}
    </td>
  );
};
