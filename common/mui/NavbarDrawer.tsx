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
import { match, useLocation } from 'react-router';
import * as H from 'history';
import AppBar from '@mui/material/AppBar';
import Drawer from '@mui/material/Drawer';
import IconButton from '@mui/material/IconButton';
import Button from '@mui/material/Button';
import MenuIcon from '@mui/icons-material/Menu';
import PersonIcon from '@mui/icons-material/Person';
import LoginIcon from '@mui/icons-material/Login';
import LogoutIcon from '@mui/icons-material/Logout';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import { PeridotLogo } from 'common/ui/PeridotLogo';
import Toolbar from '@mui/material/Toolbar';
import Divider from '@mui/material/Divider';
import { NavLink, Link, NavLinkProps, useRouteMatch } from 'react-router-dom';
import ListItem from '@mui/material/ListItem/ListItem';
import ListItemText from '@mui/material/ListItemText';
import { peridotTheme, primaryColor } from './theme';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import List from '@mui/material/List';
import Box from '@mui/material/Box';
import ListItemButton from '@mui/material/ListItemButton';
import useMediaQuery from '@mui/material/useMediaQuery';

export interface NavbarLink extends Pick<NavLinkProps, 'isActive'> {
  text: string;
  href: string;
  real?: boolean;
  exact?: boolean;

  icon?(classes: string): React.ReactNode;
}

export interface NavbarCategories {
  title?: string;
  links: (NavbarLink | undefined)[];
}

export interface NavbarDrawerProps {
  mainLinks?: NavbarCategories[];
  afterLogoNode?: () => React.ReactNode;

  logo(classes: string): React.ReactNode;
}

interface LinkRealWrapperProps extends Pick<NavLinkProps, 'isActive'> {
  to: string;
  children: React.ReactNode;

  real?: boolean;
  exact?: boolean;

  onClick?(): void;
}

const LinkRealWrapper = (props: LinkRealWrapperProps) => {
  const linkClasses = 'hover:bg-gray-100 focus:bg-gray-100';

  return props.real ? (
    <a className={linkClasses} href={props.to} onClick={props.onClick}>
      {props.children}
    </a>
  ) : (
    <NavLink
      className={linkClasses}
      activeClassName="text-peridot-primary"
      to={props.to}
      isActive={props.isActive}
      exact={props.exact}
      onClick={props.onClick}
    >
      {props.children}
    </NavLink>
  );
};

const itemNoHover = {
  py: '2px',
  px: 3,
  color: 'rgba(255, 255, 255, 0.7)',
};

const item = Object.assign({}, itemNoHover, {
  '&:hover, &:focus': {
    bgcolor: 'rgba(255, 255, 255, 0.08)',
  },
});

const itemCategory = {
  boxShadow: '0 -1px 0 rgb(255,255,255,0.1) inset',
  py: 1.5,
  px: 3,
};

export const NavbarDrawer = (props: NavbarDrawerProps) => {
  const location = useLocation();
  const [drawerOpen, setDrawerOpen] = React.useState(false);
  const isSmUp = useMediaQuery(peridotTheme.breakpoints.up('sm'));

  const toggleDrawer = () => {
    setDrawerOpen(!drawerOpen);
  };

  return (
    <>
      <Box component="nav" sx={{ width: { sm: 256 }, flexShrink: { sm: 0 } }}>
        <Drawer
          variant={isSmUp ? 'permanent' : 'temporary'}
          open={isSmUp ? true : drawerOpen}
          onClose={isSmUp ? undefined : toggleDrawer}
          PaperProps={{ style: { width: 256 } }}
          sx={isSmUp ? { display: { sm: 'block', xs: 'none' } } : null}
        >
          <List disablePadding>
            <ListItem
              sx={{
                ...itemNoHover,
                ...itemCategory,
                fontSize: 22,
                color: '#fff',
              }}
            >
              {props.logo('h-8')}
            </ListItem>
            {window.state.email ? (
              <ListItem sx={{ ...itemNoHover, ...itemCategory }}>
                <ListItemIcon>
                  <PersonIcon />
                </ListItemIcon>
                <ListItemText className="text-ellipsis overflow-hidden">
                  {window.state.name && window.state.name.length > 0
                    ? window.state.name
                    : window.state.email}
                </ListItemText>
                <a href="/oauth2/logout">
                  <LogoutIcon fontSize="small" />
                </a>
              </ListItem>
            ) : (
              <ListItem sx={{ ...itemNoHover, ...itemCategory }}>
                <a className="flex items-center" href="/oauth2/login">
                  <ListItemIcon>
                    <LoginIcon />
                  </ListItemIcon>
                  <ListItemText>Login</ListItemText>
                </a>
              </ListItem>
            )}
            {props.afterLogoNode && (
              <ListItem sx={{ ...itemNoHover, ...itemCategory }}>
                <ListItemText>{props.afterLogoNode()}</ListItemText>
              </ListItem>
            )}
            {props.mainLinks?.map((category) => (
              <Box key={category.title || 'empty'} sx={{ bgcolor: '#101F33' }}>
                {category.title ? (
                  <ListItem sx={{ py: 2, px: 3 }}>
                    <ListItemText sx={{ color: '#fff' }}>
                      {category.title}
                    </ListItemText>
                  </ListItem>
                ) : (
                  <div className="pt-4" />
                )}
                {category.links.map((link: NavbarLink | undefined) => {
                  if (!link) {
                    return null;
                  }
                  let match = useRouteMatch({
                    path: link.href,
                    exact: link.exact,
                  });
                  if (link.isActive && !link.isActive(match, location)) {
                    match = null;
                  }

                  return (
                    <LinkRealWrapper
                      key={link.href}
                      to={link.href}
                      real={link.real}
                      isActive={link.isActive}
                      exact={link.exact}
                    >
                      <ListItem disablePadding>
                        <ListItemButton selected={!!match} sx={item}>
                          {link.icon && (
                            <ListItemIcon>{link.icon('')}</ListItemIcon>
                          )}
                          <ListItemText primary={link.text} />
                        </ListItemButton>
                      </ListItem>
                    </LinkRealWrapper>
                  );
                })}
                <Divider sx={{ mt: 2 }} />
              </Box>
            ))}
          </List>
        </Drawer>
      </Box>
    </>
  );
};
