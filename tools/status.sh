#
# Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
# Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
# Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
#
# 1. Redistributions of source code must retain the above copyright notice,
# this list of conditions and the following disclaimer.
#
# 2. Redistributions in binary form must reproduce the above copyright notice,
# this list of conditions and the following disclaimer in the documentation
# and/or other materials provided with the distribution.
#
# 3. Neither the name of the copyright holder nor the names of its contributors
# may be used to endorse or promote products derived from this software without
# specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.
#

cat<<EOF
STABLE_BUILD_TAG ${GIT_COMMIT:-$(git rev-parse HEAD)}
STABLE_STAGE ${STABLE_STAGE:--dev}
STABLE_LOCAL_ENVIRONMENT ${STABLE_LOCAL_ENVIRONMENT:-0}
STABLE_DOMAIN_USER ${DOMAIN_USER:-"user-orig"}
STABLE_OCI_REGISTRY ${STABLE_OCI_REGISTRY:-host.docker.internal.local:5000}
STABLE_OCI_REGISTRY_REPO ${STABLE_OCI_REGISTRY_REPO:-dev}
STABLE_OCI_REGISTRY_DOCKER ${STABLE_OCI_REGISTRY_DOCKER:-docker.io}
STABLE_REGISTRY_SECRET ${STABLE_REGISTRY_SECRET:-none}
STABLE_OCI_REGISTRY_NO_NESTED_SUPPORT_IN_2022_SHAME_ON_YOU_AWS ${STABLE_OCI_REGISTRY_NO_NESTED_SUPPORT_IN_2022_SHAME_ON_YOU_AWS:-false}
STABLE_SITE ${STABLE_SITE:-normal}
EOF
