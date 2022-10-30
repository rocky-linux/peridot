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

insert into short_codes (code, mode)
values ('RL', 2);
insert into short_codes (code, mode)
values ('RK', 1);
--insert into products (id, name, current_full_version, redhat_major_version, short_code_code, archs, mirror_from_date, redhat_product_prefix, cpe)
--values (1, 'Rocky Linux 9', '9.0', 9, 'RL', array ['x86_64', 'aarch64', 'ppc64le', 's390x'], '2022-05-15'::timestamp, 'Rocky Linux', 'cpe:/o:redhat:enterprise_linux:9');
insert into products (id, name, current_full_version, redhat_major_version, short_code_code, archs, mirror_from_date, redhat_product_prefix, cpe, build_system, build_system_endpoint, koji_compose, koji_module_compose)
values (2, 'Rocky Linux 8', '8.6', 8, 'RL', array ['x86_64', 'aarch64'], '2022-05-15'::timestamp, 'Rocky Linux', 'cpe:/o:redhat:enterprise_linux:8', 'koji', 'https://koji.rockylinux.org/kojihub', 'dist-rocky8-compose', 'dist-rocky8-module-compose');
--insert into ignored_upstream_packages (product_id, package)
--values
--    (1, 'tfm-rubygem-unicode*'),
--    (1, 'katello-host-tools*'),
--    (1, 'openssl-ibmca*'),
--    (1, 'insights-client*'),
--    (1, 'tfm-rubygem-unicode-display_width*'),
--    (1, 'pulp*'),
--    (1, 'satellite*'),
--    (1, 'tfm-rubygem-unf_ext*'),
--    (1, 'foreman*'),
--    (1, 'kpatch*'),
--    (1, 'rhc-worker-playbook*');
insert into ignored_upstream_packages (product_id, package)
values
  (2, 'tfm-rubygem-unicode*'),
  (2, 'katello-host-tools*'),
  (2, 'openssl-ibmca*'),
  (2, 'insights-client*'),
  (2, 'tfm-rubygem-unicode-display_width*'),
  (2, 'pulp*'),
  (2, 'satellite*'),
  (2, 'tfm-rubygem-unf_ext*'),
  (2, 'foreman*'),
  (2, 'kpatch*'),
  (2, 'rhc-worker-playbook*');
insert into reboot_suggested_packages (name)
values
  ('kernel'),
  ('kernel-PAE'),
  ('kernel-rt'),
  ('kernel-smp'),
  ('kernel-xen'),
  ('linux-firmware'),
  ('*-firmware-*'),
  ('dbus'),
  ('glibc'),
  ('hal'),
  ('systemd'),
  ('udev'),
  ('gnutls'),
  ('openssl-libs');
