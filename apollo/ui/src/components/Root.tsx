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

import { Box, HStack, Text, Link as ChakraLink } from '@chakra-ui/react';
import { RESFLogo } from 'common/ui/RESFLogo';
import React from 'react';
import { Route, Switch } from 'react-router';
import { Link } from 'react-router-dom';

import { COLOR_RESF_BLUE, COLOR_RESF_GREEN } from '../styles';
import { Overview } from './Overview';
import { ShowErrata } from './ShowErrata';

export const Root = () => {
  return (
    <Box
      display="flex"
      width="100%"
      minHeight="100vh"
      flexDirection="column"
      alignItems="stretch"
    >
      <Box
        background={`linear-gradient(to right, ${COLOR_RESF_GREEN}, ${COLOR_RESF_BLUE})`}
        display="flex"
        flexDirection="row"
        alignItems="center"
        py="1"
        px={4}
      >
        <Link to="/" className="no-underline text-white">
          <HStack flexGrow={1} height="90%" spacing="2">
            <RESFLogo className="fill-current text-white" />
            <Text
              borderLeft="1px solid"
              pl="2"
              lineHeight="30px"
              fontSize="xl"
              fontWeight="300"
              color="white"
            >
              Product Errata
            </Text>
          </HStack>
        </Link>
      </Box>
      <Box as="main" flexGrow={1} overflow="auto">
        <Switch>
          <Route path="/" exact component={Overview} />
          <Route path="/:id" component={ShowErrata} />
        </Switch>
      </Box>
      <Box
        py="6"
        px="4"
        backgroundColor="gray.700"
        color="white"
        display="flex"
      >
        <ChakraLink
          href="/api/v2/advisories:rss"
          isExternal
          display="flex"
          alignItems="center"
        >
          <Box
            as="svg"
            viewBox="0 0 24 24"
            width="18px"
            height="18px"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            display="block"
            mr={1.5}
          >
            <path d="M4 11a9 9 0 019 9M4 4a16 16 0 0116 16" />
            <circle cx="5" cy="19" r="1" />
          </Box>
          <span>RSS</span>
        </ChakraLink>
      </Box>
    </Box>
  );
};
