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

// Global imports
const path = require('path');
const fs = require('fs');
const webpack = require('webpack');
const glob = require('glob');
const mergeDirs = require('merge-dirs');

const TerserJSPlugin = require('terser-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const OptimizeCSSAssetsPlugin = require('optimize-css-assets-webpack-plugin');
const CompressionPlugin = require('compression-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ReactRefreshWebpackPlugin = require('@pmmmwh/react-refresh-webpack-plugin');

// Function imports
const resolve = path.resolve; // eslint-disable-line

// Env variables
const dev = process.env.NODE_ENV !== 'production';

const bannerFile = TMPL_banner_file;
const stampData = TMPL_stamp_data;
const stableStampData = TMPL_stable_stamp_data;
const title = 'TMPL_title';
const moduleMappings = TMPL_module_mappings;
const rawName = 'TMPL_global_name';
const name = rawName
  .replace(/-/g, '_')
  .replace('.bundle', '')
  .replace('.server', '');
const indexHtml = 'TMPL_indexHtml';
const inputs = [TMPL_inputs];
const bodyScript = 'TMPL_body_script';
const headStyle = 'TMPL_head_style';
const typekit = 'TMPL_typekit';
const noSuffixFrontend = TMPL_no_suffix_frontend;

Object.keys(moduleMappings).forEach((k) => {
  moduleMappings[k] = resolve(process.cwd(), moduleMappings[k]);
});

// Entries
const entries = inputs.map((n) => resolve(process.cwd(), n));
const devEntries = dev ? ['webpack-hot-middleware/client'] : [];

const modules = [
  resolve(process.cwd(), '../npm/node_modules'),
  resolve(process.cwd(), 'external/npm/node_modules'),
  resolve(process.cwd()),
];

const volatileStatus = {};
if (!dev && stampData) {
  const versionTag = fs.readFileSync(stampData, { encoding: 'utf-8' });
  const vsSplit = versionTag.split('\n');
  for (const vss of vsSplit) {
    const vskp = vss.split(' ');
    volatileStatus[vskp[0]] = vskp[1];
  }
}

const stableStatus = {};
if (!dev && stableStampData) {
  const versionTag = fs.readFileSync(stableStampData, { encoding: 'utf-8' });
  const vsSplit = versionTag.split('\n');
  for (const vss of vsSplit) {
    const vskp = vss.split(' ');
    stableStatus[vskp[0]] = vskp[1];
  }
}

const fileName = `${name}.bundle${dev ? '' : '.min'}`;

const stage = stableStatus['STABLE_STAGE'] || '-qa';
const stageNoDash = stage.replace(/^-/, '');
const ns = `${rawName}${noSuffixFrontend ? '' : '-frontend'}`;

const apiUrl = process.env.API_URL;
const apiKey = process.env.API_KEY;

const htmlFile = resolve(process.cwd(), indexHtml);

// Plugins
// (dev)
const devPlugins = dev
  ? [new webpack.HotModuleReplacementPlugin(), new ReactRefreshWebpackPlugin()]
  : [];

// (prod)
const prodPlugins = !dev
  ? [
      new MiniCssExtractPlugin({
        filename: `${fileName}.[hash].[id].css`,
        chunkFilename: `${fileName}.[id].chunk.[hash].css`,
      }),
      new CompressionPlugin(),
    ]
  : [];

// In prod, openapi sometimes misplaces generated files, let's merge that
if (!dev) {
  const files = glob.sync('bazel-out/*-fastbuild/bin', {});
  if (files.length === 1) {
    mergeDirs.default(files[0], process.cwd(), 'skip');
  }
}

const webpackConfig = {
  devtool: dev ? 'eval-cheap-module-source-map' : undefined,

  mode: dev ? 'development' : 'production',

  entry: {
    main: [...devEntries, ...entries],
  },

  output: {
    filename: `${fileName}.[hash].[id].js`,
    chunkFilename: `${fileName}.[id].chunk.[hash].js`,
    publicPath: '/',
    library: name || 'PRD',
  },

  devServer: dev
    ? {
        publicPath: '/',
        http2: true,
        hot: true,
        liveReload: false,
        historyApiFallback: {
          index: '/',
        },
        watchOptions: {
          ignored: /node_modules/,
        },
        stats: {
          colors: true,
          hash: false,
          version: false,
          timings: false,
          assets: false,
          chunks: false,
          modules: false,
          reasons: false,
          children: false,
          source: false,
          errors: false,
          errorDetails: false,
          warnings: false,
          publicPath: false,
        },
      }
    : {},

  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx', '.less', '.scss', '.css'],
    alias: {
      'bazel-bin': resolve(process.cwd()),
      common: resolve(process.cwd(), 'common'),
      tailwind: resolve(process.cwd(), 'tailwind'),
    },
    modules,
  },

  optimization: {
    minimizer: [new OptimizeCSSAssetsPlugin(), new TerserJSPlugin()],
    splitChunks: {
      chunks: 'all',
      name: dev ? false : 'all',
      minSize: 30000,
      maxSize: 300000,
      minChunks: 1,
      maxAsyncRequests: 5,
      maxInitialRequests: 3,
      automaticNameDelimiter: '~',
      cacheGroups: {
        styles: {
          name: 'styles',
          test: /\.css$/,
          chunks: 'all',
        },
        default: {
          minChunks: 2,
          priority: -20,
          reuseExistingChunk: true,
        },
      },
    },
    runtimeChunk: 'single',
    usedExports: true,
  },

  plugins: [
    indexHtml !== 'null' &&
      new HtmlWebpackPlugin({
        filename: dev ? 'index.html' : 'index.hbs',
        template: htmlFile,
        templateParameters: {
          script: bodyScript !== '' ? bodyScript : null,
          style: headStyle !== '' ? headStyle : null,
          title: title || name || 'Peridot',
          typekit: typekit !== 'none' ? typekit : null,
        },
      }),
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV),
      'process.env.API_URL': JSON.stringify(apiUrl),
      'process.env.API_KEY': JSON.stringify(apiKey),
      'process.env.BYC_ENV': JSON.stringify(stageNoDash),
      'process.env.STABLE_STAGE': JSON.stringify(stage),
    }),
    ...devPlugins,
    ...prodPlugins,
  ],

  module: {
    rules: [
      {
        test: /\.(js|ts)x?$/,
        use: [
          {
            loader: require.resolve('babel-loader'),
            options: {
              configFile: resolve(
                process.cwd(),
                'rules_byc/internal/byc_bundle/babel.config.js'
              ),
              plugins: [dev && require.resolve('react-refresh/babel')].filter(
                Boolean
              ),
            },
          },
        ],
        exclude: /node_modules\/(?!antd)/,
        sideEffects: false,
      },
      {
        test: /\.(css|scss|sass)$/,
        use: [
          dev ? require.resolve('style-loader') : MiniCssExtractPlugin.loader,
          require.resolve('css-loader'),
          {
            loader: require.resolve('postcss-loader'),
            options: {
              postcssOptions: {
                ident: 'postcss',
                plugins: [
                  [
                    require.resolve('tailwindcss'),
                    {
                      config: resolve(process.cwd(), 'TMPL_tailwind_config'),
                    },
                  ],
                  require.resolve('autoprefixer'),
                ],
              },
            },
          },
        ],
      },
      {
        test: /\.(woff|woff2)$/,
        use: [
          {
            loader: require.resolve('file-loader'),
            options: {},
          },
        ],
      },
    ],
  },
};

module.exports = webpackConfig;
