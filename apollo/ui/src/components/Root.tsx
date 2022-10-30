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
  AppBar,
  Toolbar,
  Container,
  CssBaseline,
  Drawer,
  Divider,
  IconButton,
  List,
} from '@mui/material';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import MenuIcon from '@mui/icons-material/Menu';

import { useStyles } from '../styles';
import { Switch, Route } from 'react-router';
import { Overview } from './Overview';
import { Link } from 'react-router-dom';
import { RESFLogo } from 'common/ui/RESFLogo';
import classnames from 'classnames';
import { ShowErrata } from './ShowErrata';

export const Root = () => {
  const [open, setOpen] = React.useState(false);
  const classes = useStyles();

  const handleDrawerClose = () => {
    setOpen(false);
  };

  const handleDrawerOpen = () => {
    setOpen(true);
  };

  const inManage = location.pathname.startsWith('/manage');

  return (
    <div className={classes.root}>
      <AppBar
        position="absolute"
        className={classnames(
          inManage && classes.appBar,
          open && classes.appBarShift
        )}
      >
        <Toolbar className={classes.toolbar}>
          {inManage && (
            <IconButton
              edge="start"
              color="inherit"
              aria-label="open drawer"
              onClick={handleDrawerOpen}
              className={classnames(
                classes.menuButton,
                open && classes.menuButtonHidden
              )}
            >
              <MenuIcon />
            </IconButton>
          )}
          <Link to="/" className="no-underline text-white">
            <div
              className={classnames(
                classes.title,
                'flex items-center space-x-4'
              )}
            >
              <RESFLogo className="fill-current text-white" />
              <div className="font-bold text-lg text-white no-underline">
                Product Errata{inManage && ' (Admin)'}
              </div>
            </div>
          </Link>
        </Toolbar>
      </AppBar>
      {inManage && (
        <Drawer
          variant="permanent"
          classes={{
            paper: classnames(
              classes.drawerPaper,
              !open && classes.drawerPaperClose
            ),
          }}
          open={open}
        >
          <div className={classes.toolbarIcon}>
            <IconButton onClick={handleDrawerClose}>
              <ChevronLeftIcon />
            </IconButton>
          </div>
          <Divider />
        </Drawer>
      )}
      <main className={classes.content}>
        <div className={classes.appBarSpacer} />
        <Switch>
          <Route path="/" exact component={Overview} />
          <Route path="/:id" component={ShowErrata} />
        </Switch>
      </main>
    </div>
  );
};
