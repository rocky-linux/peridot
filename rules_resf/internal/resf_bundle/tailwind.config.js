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

const path = require('path');

const defFonts = [
  'Inter',
  '-apple-system',
  'BlinkMacSystemFont',
  '"Segoe UI"',
  'Roboto',
  '"Helvetica Neue"',
  'Arial',
  '"Noto Sans"',
  'sans-serif',
  '"Apple Color Emoji"',
  '"Segoe UI Emoji"',
  '"Segoe UI Symbol"',
  '"Noto Color Emoji"',
];

const fontSize = {
  xxs: '0.55rem',
  xs: '0.75rem',
  lxs: '0.8rem',
  sm: '0.875rem',
  base: '1rem',
  lg: '1.125rem',
  xl: '1.25rem',
  '2xl': '1.5rem',
  '3xl': '1.875rem',
  '4xl': '2.25rem',
  '5xl': '3rem',
  '6xl': '4rem',
};

const rootDir = path.resolve(process.cwd());

let projectDir = rootDir;
let projectPath = rootDir;

module.exports = {
  important: true,
  mode: 'jit',
  purge: [
    path.join(projectPath, '**/*.{jsx,tsx,vue}'),
    path.join(projectPath, '../rules_resf/internal/resf_bundle/*.hbs'),
    path.resolve(path.join('.', projectDir, '**/*.{jsx,tsx,vue}')),
    path.resolve('./rules_resf/internal/resf_bundle/*.hbs'),
  ],
  plugins: [require('@tailwindcss/forms')],
  theme: {
    fontFamily: {
      sans: defFonts,
      alanding: defFonts,
    },
    fontSize,
    inset: {
      '-gone': '-4000%',
      '-full': '-100%',
      '-55p': '-55%',
      '-37p': '-37%',
      0: '0',
      1: '1rem',
      2: '2rem',
      4: '4rem',
    },
    extend: {
      colors: {
        peridot: {
          primary: '#009be5',
        },
        primary: {
          1: '#182026',
          2: '#1d272f',
        },
        blue: {
          50: '#F2F7FF',
          100: '#E6F0FF',
          200: '#BFD9FF',
          300: '#99C2FF',
          400: '#4D94FF',
          500: '#0066FF',
          600: '#005CE6',
          700: '#003D99',
          800: '#002E73',
          900: '#001F4D',
        },
        purple: {
          500: '#4f0080',
        },
      },
      boxShadow: {
        subtle: '0 0 10px 0 rgba(0, 0, 0, 0.05)',
        'subtle-lg': '0 0 10px 0 rgba(0, 0, 0, 0.07)',
        'subtle-sm-error': '0 0 2px 0 rgba(254, 178, 178, 1)',
        'subtle-xl': '0 0 10px 0 rgba(0, 0, 0, 0.1)',
      },
      width: {
        '1/7': '14.2857143%',
        '1/8': '12.5%',
        '1/9': '11.111%',
      },
      zIndex: {
        '-1': '-1',
      },
    },
  },
  variants: {
    borderStyle: ['responsive', 'hover', 'focus', 'focus-within'],
    textColor: ['responsive', 'hover', 'focus', 'group-hover'],
    boxShadow: ['responsive', 'hover', 'focus', 'group-hover'],
    display: ['responsive', 'hover', 'focus', 'group-hover'],
    height: ['responsive', 'hover', 'focus', 'group-hover'],
    backgroundColor: [
      'responsive',
      'hover',
      'focus',
      'group-hover',
      'focus-within',
    ],
    opacity: ['responsive', 'hover', 'focus', 'group-hover', 'focus-within'],
    pointerEvents: [
      'responsive',
      'hover',
      'focus',
      'group-hover',
      'focus-within',
    ],
    inset: ['responsive', 'hover', 'focus', 'group-hover', 'focus-within'],
  },
};
