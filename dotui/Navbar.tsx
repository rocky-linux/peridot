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
import { Link, NavLink } from 'react-router-dom';
import { TextField } from 'dotui/TextField';

export interface NavbarLink {
  text: string;
  href: string;
}

export interface NavbarProps {
  mainLinks?: NavbarLink[];
  subtitle?: string;
  rootLink?: string;
  subtitleLink?: string;

  logo(classes: string): React.ReactNode;
}

export const Navbar = (props: NavbarProps) => {
  return (
    <div className="w-screen bg-white flex item-start justify-between">
      <div className="flex items-center">
        <div id="brand" className="p-2 px-4 flex">
          <Link to={props.rootLink || '/'}>{props.logo('h-8 w-32')}</Link>
          {props.subtitle && (
            <a
              id="subtitle"
              className="ml-4 text-sm font-light flex items-center"
              href={props.subtitleLink || '#'}
            >
              {props.subtitle}
            </a>
          )}
        </div>
        <div id="navigation" className="flex items-center space-x-2">
          {props.mainLinks?.map((link) => (
            <NavLink
              className="flex items-center text-sm py-1 px-3 rounded hover:bg-gray-100 focus:bg-gray-100"
              activeClassName="bg-gray-100"
              to={`${props.rootLink || ''}${link.href}`}
            >
              {link.text}
            </NavLink>
          ))}
        </div>
      </div>
      <div
        id="search"
        className="flex items-center justify-center ml-auto mr-4"
      >
        <TextField placeholder label="Search" id="search" />
      </div>
      <div id="user" className="flex items-center justify-end mr-4 text-sm">
        Mustafa Gezen
      </div>
    </div>
  );
};
