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

import express from 'express';
import httpProxyMiddleware from 'http-proxy-middleware';
import bodyParser from 'body-parser';
import cookieParser from 'cookie-parser';
import helmet from 'helmet';
import webpack from 'webpack';
import webpackDevMiddleware from 'webpack-dev-middleware';
import webpackHotMiddleware from 'webpack-hot-middleware';
import webpackMildCompile from 'webpack-mild-compile';
import expressOidc from 'express-openid-connect';
import history from 'connect-history-api-fallback';
import hbs from 'hbs';
import evilDns from 'evil-dns';
import fs from 'fs';
import dns from 'dns';

const { createProxyMiddleware } = httpProxyMiddleware;

const { auth } = expressOidc;

export default async function (opts) {
  // Create a new app for health checks.
  const appZ = express();
  appZ.get('/healthz', (req, res) => {
    res.end();
  });
  appZ.get('/_/healthz', (req, res) => {
    res.end();
  });

  const app = express();
  app.use(function (req, res, next) {
    // Including resf-internal-req: 1 should return the Z page
    if (req.header('resf-internal-req') === 'yes') {
      appZ(req, res, next);
    } else {
      next();
    }
  });
  const prod = process.env.NODE_ENV === 'production';

  const port = prod ? process.env.PORT || 8086 : opts.port;

  opts.secret = process.env.RESF_SECRET;

  // If we're in prod, then a secret has to be present
  if (prod && (!opts.secret || opts.secret.length < 32)) {
    throw 'secret has to be at least 32 characters';
  }

  // Add authentication if not disabled
  if (!opts.disableAuth) {
    console.log(`Using issuer: ${opts.issuerBaseURL}`);
    console.log(`Using clientID: ${opts.clientID}`);
    console.log(`Using baseURL: ${opts.baseURL}`);

    if (
      (opts.issuerBaseURL.endsWith('.localhost') ||
        opts.issuerBaseURL.endsWith('.localhost/')) &&
      process.env['RESF_ENV']
    ) {
      const kong = 'istio-ingressgateway.istio-system.svc.cluster.local';
      const urlObject = new URL(opts.issuerBaseURL);
      console.warn(`Forcing ${urlObject.hostname} to resolve to ${kong}`);
      const lookup = async () => {
        return new Promise((resolve, reject) => {
          // noinspection HttpUrlsUsage
          dns.lookup(kong, { family: 4 }, (err, address, family) => {
            if (err) {
              reject(err);
            }
            resolve(address);
          });
        });
      };
      const internalServiceResolve = await lookup();
      evilDns.add(urlObject.hostname, internalServiceResolve);
      // Disable TLS verification for development
      process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = 0;
    }

    const config = {
      authRequired: process.env['DISABLE_AUTH_ENFORCE']
        ? process.env['DISABLE_AUTH_ENFORCE'] === 'false'
        : !!!opts.disableAuthEnforce,
      // Disable telemetry
      enableTelemetry: false,
      // Use dev secret is none is present (Prod requires a secret so not a security issue)
      secret: opts.secret || 'dev-secret-123',
      // Add BaseURL for callback purposes. This has to be specified in the initial server call
      // The FRONTEND_URL environment variable can override this value in prod.
      baseURL: opts.baseURL,
      // The specific application should supply a dev client ID while prod IDs should be set as an env variable
      clientID: opts.clientID,
      // The specific application should supply a dev secret while prod secrets should be set as an env variable
      clientSecret: opts.clientSecret,
      issuerBaseURL: opts.issuerBaseURL,
      idpLogout: true,
      authorizationParams: {
        response_type: 'code',
        scope: 'openid profile email offline_access',
      },
      session: {
        rolling: true,
        rollingDuration: 86400,
        absoluteDuration: 86400 * 7,
      },
      routes: {
        callback: '/oauth2/callback',
        logout: '/oauth2/logout',
        login: '/oauth2/login',
      },
    };

    // If we have a authentication prefix, only force redirect on paths with that prefix
    // Remember, authentication done here is only for simplicity purposes.
    // The authentication token is then passed on to the API.
    // Bypassing auth here doesn't accomplish anything.
    let middlewares = [];

    // If requireEmailSuffix is present, let's validate post callback
    // that the signed in email ends with a suffix in the allowlist
    // Again, a bypass here doesn't accomplish anything.
    let requireEmailSuffix = opts.authOptions?.requireEmailSuffix;
    if (process.env['AUTH_OPTIONS_REQUIRE_EMAIL_SUFFIX']) {
      requireEmailSuffix =
        process.env['AUTH_OPTIONS_REQUIRE_EMAIL_SUFFIX'].split(',');
    }
    if (requireEmailSuffix) {
      middlewares.push((req, res, next) => {
        const email = req.oidc?.user?.email;
        if (!email) {
          return next('No email found in the user object');
        }
        const suffixes = requireEmailSuffix;
        let isAllowed = false;
        for (const suffix of suffixes) {
          if (email.endsWith(suffix)) {
            isAllowed = true;
            break;
          }
        }

        if (isAllowed) {
          next();
        } else {
          res.redirect(
            process.env['AUTH_REJECT_REDIRECT_URL']
              ? process.env['AUTH_REJECT_REDIRECT_URL']
              : opts.authOptions.authRejectRedirectUrl ||
                  'https://rockylinux.org'
          );
        }
      });
    }

    app.use(
      (req, res, next) => {
        try {
          auth(config)(req, res, next);
        } catch (err) {
          next(err);
        }
      },
      [middlewares]
    );
  }

  // Currently in dev, webpack is handling all file serving
  // This is just a placeholder
  let distDir = process.cwd() + '/dist';
  if (prod) {
    // Enable security hardening in prod
    app.use(
      helmet({
        contentSecurityPolicy: false,
      })
    );

    // Prod expects a certain container structure for all apps
    // Packaging this application with the web base should do
    // all this for you
    const dirs = fs.readdirSync('/home/app/bundle');
    distDir = `/home/app/bundle/${dirs[0]}`;
  }
  app.set('views', distDir);
  app.use(cookieParser());
  app.set('view engine', 'hbs');
  // Use the handlebar engine
  app.engine('hbs', hbs.__express);

  app.use(express.static(distDir));

  if (opts.apis) {
    Object.keys(opts.apis).forEach((x) => {
      app.use(x, async (req, res, next) => {
        let authorization = '';

        // If we have an authenticated user, send the token with the request
        if (req.oidc && req.oidc.accessToken) {
          let { access_token, isExpired, refresh } = req.oidc.accessToken;
          if (isExpired()) {
            try {
              ({ access_token } = await refresh());
            } catch (err) {
              res.oidc.logout({ returnTo: '/' });
              return next('User has to re-authenticate');
            }
          }
          authorization = `Bearer ${access_token}`;

          if (!prod) {
            console.log(`Using id token: ${req.oidc.idToken}`);
            console.log(`Using access token: ${access_token}`);
          }
        }

        const rewrite = {};
        rewrite[`^${x}`] = '';

        // Make it possible to override api url using an env variable.
        // Example: /api can be set with URL_API
        // Example 2: /manage/api can be set with URL_MANAGE_API
        const prodEnvName = `URL_${x
          .substr(1)
          .replace('/', '_')
          .toUpperCase()}`;

        const apiUrl = process.env[prodEnvName]
          ? process.env[prodEnvName]
          : prod
          ? opts.apis[x].prodApiUrl
          : opts.apis[x].devApiUrl;

        createProxyMiddleware({
          target: apiUrl,
          changeOrigin: true,
          headers: {
            host: apiUrl,
            authorization,
          },
          pathRewrite: rewrite,
        })(req, res);
      });
    });
  }

  // Template parameters for values in initial state
  const templateParams = (req) => {
    // If auth is disabled, then either return an empty list
    // or run the templateFunc WITHOUT a user.
    // It's important that apps do not use `user` without validation
    if (opts.disableAuth || !req.oidc) {
      if (opts.templateFunc) {
        return opts.templateFunc();
      }
      return {};
    }

    const { user } = req.oidc;
    if (!user) {
      return {};
    }

    if (opts.templateFunc) {
      return opts.templateFunc(user);
    }

    // Return default values
    const { email, name, picture } = user;
    return {
      email,
      name,
      picture,
    };
  };

  if (prod) {
    app.get('/*', (req, res) => {
      // Prod doesn't do hacky shit with the webpack compiler so just add the params
      // to the locals
      res.locals = templateParams(req);

      res.render('index');
    });
  } else {
    // Here comes the hack train
    if (!opts.webpackConfig && opts.webpackPath) {
      opts.webpackConfig = await import(opts.webpackPath);
    }
    // Create a live-reloading dev instance of the app with the given webpack config
    const compiler = webpack(opts.webpackConfig);
    webpackMildCompile(compiler);

    const wdm = webpackDevMiddleware(compiler, {
      publicPath: opts.webpackConfig.output.publicPath,
    });

    app.use(history());
    app.use((req, res, next) => {
      // Here we cache the old send function to re-use after we run the HTML through handlebars
      const oldSend = res.send;

      res.send = (data) => {
        let newData;
        // Check if the request returned a HTML page
        // For SPAs, the only HTML page is the index page
        if (res.get('content-type').indexOf('text/html') !== -1) {
          // Run through handlebars compiler with our template parameters
          newData = hbs.handlebars.compile(data.toString())(
            templateParams(req)
          );
        } else {
          // No new data, just return old data
          newData = data;
        }
        // Re-replace res.send with the old res.send
        res.send = oldSend;
        // Run the old res.send with the new data
        return res.send(newData);
      };

      next();
    });
    // Enable hot reload
    app.use(wdm);
    app.use(webpackHotMiddleware(compiler));
  }

  // Enable JSON bodies. We're forwarding this to the API
  app.use(bodyParser.json());

  console.log(`view app on ${opts.baseURL}`);
  app.listen(port);
}
