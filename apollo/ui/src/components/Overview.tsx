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
  AddIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  ChevronDownIcon,
  MinusIcon,
  SearchIcon,
} from '@chakra-ui/icons';
import {
  Alert,
  AlertDescription,
  AlertIcon,
  AlertTitle,
  Box,
  ButtonGroup,
  FormControl,
  FormLabel,
  HStack,
  IconButton,
  Input,
  InputGroup,
  InputLeftElement,
  Select,
  Spinner,
  Stack,
  Table,
  TableColumnHeaderProps,
  TableContainer,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  severityToBadge,
  severityToText,
  typeToBadge,
  typeToText,
} from 'apollo/ui/src/enumToText';
import {
  ListAdvisoriesFiltersSeverityEnum,
  ListAdvisoriesFiltersTypeEnum,
} from 'bazel-bin/apollo/proto/v1/client_typescript';
import {
  AdvisorySeverity,
  V1Advisory,
  V1AdvisoryType,
} from 'bazel-bin/apollo/proto/v1/client_typescript/models';
import { reqap } from 'common/ui/reqap';
import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import { api } from '../api';
import { COLOR_RESF_GREEN } from '../styles';

export const Overview = () => {
  const inputBackground = useColorModeValue('white', 'gray.800');

  const [advisories, setAdvisories] = useState<V1Advisory[]>();
  const [lastUpdated, setLastUpdated] = useState<Date>();
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  // Request State
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(25);
  const [filtersKeyword, setFiltersKeyword] = useState<string>();
  const [filterBefore, setFilterBefore] = useState<Date>();
  const [filterAfter, setFilterAfter] = useState<Date>();
  const [filtersType, setFiltersType] =
    useState<keyof typeof ListAdvisoriesFiltersTypeEnum>();
  const [filtersSeverity, setFiltersSeverity] =
    useState<keyof typeof ListAdvisoriesFiltersSeverityEnum>();

  useEffect(() => {
    const fetch = async () => {
      setIsLoading(true);
      const [err, res] = await reqap(() =>
        api.listAdvisories({
          page,
          limit: pageSize,
          filtersKeyword,
          filtersBefore: filterBefore,
          filtersAfter: filterAfter,
          filtersSeverity: filtersSeverity
            ? ListAdvisoriesFiltersSeverityEnum[filtersSeverity]
            : undefined,
          filtersType: filtersType
            ? ListAdvisoriesFiltersTypeEnum[filtersType]
            : undefined,
        })
      );

      setIsLoading(false);

      if (err || !res) {
        setIsError(true);
        setAdvisories(undefined);
        return;
      }

      setIsError(false);

      if (res) {
        setAdvisories(res.advisories);
        setLastUpdated(res.lastUpdated);
        setTotal(parseInt(res.total || '0'));
      }
    };

    const timer = setTimeout(() => fetch(), 500);

    return () => clearTimeout(timer);
  }, [
    pageSize,
    page,
    filtersKeyword,
    filterBefore,
    filterAfter,
    filtersSeverity,
    filtersType,
  ]);

  // TODO: Figure out why sticky isn't sticking
  const stickyProps: TableColumnHeaderProps = {
    position: 'sticky',
    top: '0px',
    zIndex: '10',
    scope: 'col',
  };

  const lastPage = total < pageSize ? 0 : Math.ceil(total / pageSize) - 1;

  return (
    <Box
      w="100%"
      h="100%"
      display="flex"
      flexDirection="column"
      p={4}
      alignItems="stretch"
    >
      <Stack
        direction={{
          sm: 'column',
          lg: 'row',
        }}
        alignItems={{
          sm: 'stretch',
          lg: 'flex-end',
        }}
      >
        <InputGroup>
          <InputLeftElement>
            <SearchIcon />
          </InputLeftElement>
          <Input
            type="search"
            aria-label="Keyword search"
            placeholder="Keyword Search"
            flexGrow={1}
            width="200px"
            variant="filled"
            borderRadius="0"
            backgroundColor={inputBackground}
            onChange={(e) => setFiltersKeyword(e.target.value)}
          />
        </InputGroup>
        <HStack>
          <FormControl width="180px" flexShrink={0} flexGrow={1}>
            <FormLabel fontSize="sm">Type</FormLabel>
            <Select
              aria-label="Type"
              placeholder="All"
              variant="filled"
              background={inputBackground}
              borderRadius="0"
              value={filtersType}
              onChange={(e) => {
                if (e.currentTarget.value !== 'Security') {
                  setFiltersSeverity(undefined);
                }

                setFiltersType(
                  e.currentTarget
                    .value as keyof typeof ListAdvisoriesFiltersTypeEnum
                );
              }}
            >
              {Object.keys(ListAdvisoriesFiltersTypeEnum)
                .sort((a, b) => a.localeCompare(b))
                .filter((a) => a !== 'Unknown')
                .map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
            </Select>
          </FormControl>
          {filtersType === 'Security' && (
            <FormControl width="180px" flexShrink={0} flexGrow={1}>
              <FormLabel fontSize="sm">Severity</FormLabel>
              <Select
                aria-label="Severity"
                placeholder="All"
                variant="filled"
                background={inputBackground}
                borderRadius="0"
                value={filtersSeverity}
                onChange={(e) =>
                  setFiltersSeverity(
                    e.currentTarget
                      .value as keyof typeof ListAdvisoriesFiltersSeverityEnum
                  )
                }
              >
                {Object.keys(ListAdvisoriesFiltersSeverityEnum)
                  .sort((a, b) => a.localeCompare(b))
                  .filter((a) => a !== 'Unknown')
                  .map((s) => (
                    <option key={s} value={s}>
                      {s}
                    </option>
                  ))}
              </Select>
            </FormControl>
          )}
        </HStack>
        <HStack>
          <FormControl width="180px" flexShrink={0} flexGrow={1}>
            <FormLabel fontSize="sm">After</FormLabel>
            <Input
              type="date"
              variant="filled"
              background={inputBackground}
              borderRadius="0"
              max={
                filterBefore
                  ? filterBefore.toLocaleDateString('en-ca')
                  : new Date().toLocaleDateString('en-ca')
              }
              value={filterAfter?.toLocaleDateString('en-ca') || ''}
              onChange={(e) => {
                const newVal = e.currentTarget.value;
                console.log(newVal);

                if (!newVal) {
                  setFilterAfter(undefined);
                }

                const asDate = new Date(newVal);
                if (!(asDate instanceof Date) || isNaN(asDate.getTime())) {
                  // Check value parses as a date
                  return;
                }

                const [year, month, date] = newVal.split('-').map(Number);

                setFilterAfter(new Date(year, month - 1, date));
              }}
            />
          </FormControl>
          <FormControl width="180px" flexShrink={0} flexGrow={1}>
            <FormLabel fontSize="sm">Before</FormLabel>
            <Input
              type="date"
              variant="filled"
              background={inputBackground}
              borderRadius="0"
              min={filterAfter?.toLocaleDateString('en-ca')}
              max={new Date().toLocaleDateString('en-ca')}
              value={filterBefore?.toLocaleDateString('en-ca') || ''}
              onChange={(e) => {
                const newVal = e.currentTarget.value;

                if (!newVal) {
                  setFilterBefore(undefined);
                }

                const asDate = new Date(newVal);
                if (!(asDate instanceof Date) || isNaN(asDate.getTime())) {
                  // Check value parses as a date
                  return;
                }

                const [year, month, date] = newVal.split('-').map(Number);

                setFilterBefore(new Date(year, month - 1, date));
              }}
            />
          </FormControl>
        </HStack>
      </Stack>
      <HStack my={4} justifyContent="space-between" flexWrap="wrap">
        <Text fontStyle="italic" fontSize="xs">
          Last updated {lastUpdated?.toLocaleString() || 'never'}
        </Text>
        <HStack>
          <Text fontSize="xs">
            Displaying {(page * pageSize + 1).toLocaleString()}-
            {Math.min(total, page * pageSize + pageSize).toLocaleString()} of{' '}
            {total.toLocaleString()}
          </Text>
          <ButtonGroup
            size="xs"
            isAttached
            alignItems="stretch"
            colorScheme="blackAlpha"
          >
            <IconButton
              aria-label="First Page"
              icon={<ArrowLeftIcon fontSize="8px" />}
              disabled={page <= 0}
              onClick={() => setPage(0)}
            />
            <IconButton
              aria-label="Previous Page"
              icon={<MinusIcon fontSize="8px" />}
              disabled={page <= 0}
              onClick={() => setPage((old) => old - 1)}
            />
            <Text
              fontSize="xs"
              // borderTop="1px solid"
              // borderBottom="1px solid"
              borderColor="gray.200"
              backgroundColor="white"
              lineHeight="24px"
              px={2}
            >
              {(page + 1).toLocaleString()} / {(lastPage + 1).toLocaleString()}
            </Text>
            <IconButton
              aria-label="Next Page"
              icon={<AddIcon fontSize="8px" />}
              disabled={page >= lastPage}
              onClick={() => setPage((old) => old + 1)}
            />
            <IconButton
              aria-label="Last Page"
              icon={<ArrowRightIcon fontSize="8px" />}
              disabled={page >= lastPage}
              onClick={() => setPage(lastPage)}
            />
          </ButtonGroup>
        </HStack>
      </HStack>
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
        <Box backgroundColor="white" boxShadow="base">
          <TableContainer>
            <Table size="sm" variant="striped">
              <Thead>
                <Tr>
                  <Th {...stickyProps} width="36px" />
                  <Th {...stickyProps}>Advisory</Th>
                  <Th {...stickyProps}>Synopsis</Th>
                  <Th {...stickyProps}>Type / Severity</Th>
                  <Th {...stickyProps}>Products</Th>
                  <Th {...stickyProps}>
                    <HStack spacing={1}>
                      <Text>Issue Date</Text>
                      <ChevronDownIcon />
                    </HStack>
                  </Th>
                </Tr>
              </Thead>
              <Tbody>
                {!advisories?.length && (
                  <Tr>
                    <Td colSpan={6} textAlign="center">
                      <Text>No rows found</Text>
                    </Td>
                  </Tr>
                )}
                {advisories?.map((a) => (
                  <Tr key={a.name}>
                    <Td textAlign="center" pr={0}>
                      {severityToBadge(a.severity)}
                    </Td>
                    <Td>
                      <Link
                        className="text-peridot-primary visited:text-purple-500"
                        to={`/${a.name}`}
                      >
                        {a.name}
                      </Link>
                    </Td>
                    <Td>
                      {a.synopsis?.replace(
                        /^(Critical|Important|Moderate|Low): /,
                        ''
                      )}
                    </Td>
                    <Td>
                      {typeToText(a.type)}
                      {a.type === V1AdvisoryType.Security
                        ? ` / ${severityToText(a.severity)}`
                        : ''}
                    </Td>
                    <Td>{a.affectedProducts?.join(', ')}</Td>
                    <Td>
                      {Intl.DateTimeFormat(undefined, {
                        day: '2-digit',
                        month: 'short',
                        year: 'numeric',
                      }).format(a.publishedAt)}
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </TableContainer>
        </Box>
      )}
      <HStack justifyContent="flex-end" mt={4}>
        <Text as="label" htmlFor="row-count" fontSize="sm">
          Rows per page:
        </Text>
        <Select
          id="row-count"
          name="row-count"
          variant="filled"
          backgroundColor={inputBackground}
          width="100px"
          size="sm"
          value={pageSize}
          onChange={(e) => {
            setPage(0);
            setPageSize(Number(e.currentTarget.value));
          }}
        >
          {[10, 25, 50, 100].map((count) => (
            <option key={count} value={count}>
              {count.toLocaleString()}
            </option>
          ))}
        </Select>
      </HStack>
    </Box>
  );
};
