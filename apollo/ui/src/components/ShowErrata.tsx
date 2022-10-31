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

import {
  Alert,
  AlertDescription,
  AlertIcon,
  AlertTitle,
  Box,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  Heading,
  HStack,
  Link,
  ListItem,
  Spinner,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
  UnorderedList,
  VStack,
} from '@chakra-ui/react';
import { severityToBadge, typeToText } from 'apollo/ui/src/enumToText';
import { V1Advisory } from 'bazel-bin/apollo/proto/v1/client_typescript';
import { reqap } from 'common/ui/reqap';
import React, { useState } from 'react';
import { RouteComponentProps } from 'react-router';
import { Link as RouterLink } from 'react-router-dom';

import { api } from '../api';
import { COLOR_RESF_BLUE, COLOR_RESF_GREEN } from '../styles';

interface ShowErrataParams {
  id: string;
}

export interface ShowErrataProps
  extends RouteComponentProps<ShowErrataParams> {}

export const ShowErrata = (props: ShowErrataProps) => {
  const id = props.match.params.id;

  const [errata, setErrata] = useState<V1Advisory>();
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  React.useEffect(() => {
    const fetch = async () => {
      setIsLoading(true);

      const [err, res] = await reqap(() => api.getAdvisory({ id }));

      setIsLoading(false);

      if (err || !res) {
        setIsError(true);
        setErrata(undefined);
        return;
      }

      setIsError(false);

      setErrata(res.advisory);
    };

    fetch();
  }, [id]);

  return (
    <Box
      w="100%"
      h="100%"
      display="flex"
      flexDirection="column"
      p={4}
      alignItems="stretch"
      maxWidth="1300px"
      m="auto"
    >
      <Breadcrumb mb={4}>
        <BreadcrumbItem>
          <BreadcrumbLink as={RouterLink} to="/">
            Product Errata
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbItem>
          <BreadcrumbLink isCurrentPage>{id}</BreadcrumbLink>
        </BreadcrumbItem>
      </Breadcrumb>
      {isLoading ? (
        <Spinner
          m="auto"
          size="xl"
          alignSelf="center"
          color={COLOR_RESF_GREEN}
          thickness="3px"
        />
      ) : isError ? (
        <Alert
          status="error"
          m="auto"
          flexDirection="column"
          width="300px"
          borderRadius="md"
        >
          <AlertIcon mr="0" />
          <AlertTitle>Something has gone wrong</AlertTitle>
          <AlertDescription>Failed to load errata</AlertDescription>
        </Alert>
      ) : (
        errata && (
          <>
            <HStack
              alignItems="center"
              backgroundColor="white"
              py="2"
              px="4"
              spacing="6"
              mb={2}
            >
              {severityToBadge(errata.severity, 40)}
              <VStack alignItems="stretch" spacing="0" flexGrow={1}>
                <HStack justifyContent="space-between">
                  <Text fontSize="lg" fontWeight="bold">
                    {errata.name}
                  </Text>
                </HStack>
                <Text fontSize="sm">{errata.synopsis}</Text>
              </VStack>
            </HStack>
            <Tabs backgroundColor="white" p="2">
              <TabList>
                <Tab>Erratum</Tab>
                <Tab>Affected Packages</Tab>
              </TabList>
              <Box
                display="flex"
                flexDir="row"
                alignItems="stretch"
                flexWrap="wrap"
                justifyContent="space-between"
              >
                <TabPanels maxWidth="850px" px="2">
                  <TabPanel>
                    <Heading as="h2" size="md">
                      Topic
                    </Heading>
                    {errata.topic?.split('\n').map((p, i) => (
                      <Text key={i} mt={2}>
                        {p}
                      </Text>
                    ))}
                    <Heading as="h2" size="md" mt={4}>
                      Description
                    </Heading>
                    {errata.description?.split('\n').map((p, i) => (
                      <Text key={i} mt={2}>
                        {p}
                      </Text>
                    ))}
                  </TabPanel>
                  <TabPanel>
                    <VStack alignItems="flex-start" spacing="6">
                      {Object.keys(errata.rpms || {}).map((product) => (
                        <div key={product}>
                          <Heading as="h2" size="lg" mb={4} fontWeight="300">
                            {product}
                          </Heading>
                          <Heading as="h3" size="md" mt={2}>
                            SRPMs
                          </Heading>
                          <UnorderedList pl="4">
                            {errata.rpms?.[product]?.nvras
                              ?.filter((x) => x.indexOf('.src.rpm') !== -1)
                              .map((x) => (
                                <ListItem key={x}>{x}</ListItem>
                              ))}
                          </UnorderedList>
                          <Heading as="h3" size="md" mt={2}>
                            RPMs
                          </Heading>
                          <UnorderedList pl="4">
                            {errata.rpms?.[product]?.nvras
                              ?.filter((x) => x.indexOf('.src.rpm') === -1)
                              .map((x) => (
                                <ListItem key={x}>{x}</ListItem>
                              ))}
                          </UnorderedList>
                        </div>
                      ))}
                    </VStack>
                  </TabPanel>
                </TabPanels>
                <VStack
                  py="4"
                  px="8"
                  alignItems="flex-start"
                  minWidth="300px"
                  spacing="5"
                  flexShrink={0}
                  backgroundColor="gray.100"
                >
                  <Text>
                    <b>Issued:</b> {errata.publishedAt?.toLocaleDateString()}
                  </Text>
                  <Text>
                    <b>Type:</b> {typeToText(errata.type)}
                  </Text>
                  <Box>
                    <Text fontWeight="bold">
                      Affected Product
                      {(errata.affectedProducts?.length || 0) > 1 ? 's' : ''}
                    </Text>
                    <UnorderedList>
                      {errata.affectedProducts?.map((x, idx) => (
                        <ListItem key={idx}>{x}</ListItem>
                      ))}
                    </UnorderedList>
                  </Box>
                  <Box>
                    <Text fontWeight="bold">Fixes</Text>
                    <UnorderedList>
                      {errata.fixes?.map((x, idx) => (
                        <ListItem key={idx}>
                          <Link
                            href={x.sourceLink}
                            isExternal
                            color={COLOR_RESF_BLUE}
                          >
                            {x.sourceBy} - {x.ticket}
                          </Link>
                        </ListItem>
                      ))}
                    </UnorderedList>
                  </Box>
                  <Box>
                    <Text fontWeight="bold">CVEs</Text>
                    <UnorderedList>
                      {!!errata.cves?.length ? (
                        errata.cves?.map((x, idx) => {
                          let text = `${x.name}${
                            x.sourceBy !== '' && ` (Source: ${x.sourceBy})`
                          }`;

                          return (
                            <ListItem key={idx}>
                              {x.sourceLink === '' ? (
                                <span>{text}</span>
                              ) : (
                                <Link
                                  href={x.sourceLink}
                                  isExternal
                                  color={COLOR_RESF_BLUE}
                                >
                                  {text}
                                </Link>
                              )}
                            </ListItem>
                          );
                        })
                      ) : (
                        <ListItem>No CVEs</ListItem>
                      )}
                    </UnorderedList>
                  </Box>
                  <Box>
                    <Text fontWeight="bold">References</Text>
                    <UnorderedList>
                      {!!errata.references?.length ? (
                        errata.references?.map((x, idx) => (
                          <ListItem key={idx}>{x}</ListItem>
                        ))
                      ) : (
                        <ListItem>No references</ListItem>
                      )}
                    </UnorderedList>
                  </Box>
                </VStack>
              </Box>
            </Tabs>
          </>
        )
      )}
    </Box>
  );
};
