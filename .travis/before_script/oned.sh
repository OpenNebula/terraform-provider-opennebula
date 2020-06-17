# -------------------------------------------------------------------------- #
# Copyright 2002-2019, OpenNebula Project, OpenNebula Systems                #
#                                                                            #
# Licensed under the Apache License, Version 2.0 (the "License"); you may    #
# not use this file except in compliance with the License. You may obtain    #
# a copy of the License at                                                   #
#                                                                            #
# http://www.apache.org/licenses/LICENSE-2.0                                 #
#                                                                            #
# Unless required by applicable law or agreed to in writing, software        #
# distributed under the License is distributed on an "AS IS" BASIS,          #
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   #
# See the License for the specific language governing permissions and        #
# limitations under the License.                                             #
#--------------------------------------------------------------------------- #

# Set credentials
sudo su -c "echo 'oneadmin:opennebula' > /var/lib/one/.one/one_auth" oneadmin

mkdir ~/.one
echo 'oneadmin:opennebula' > ~/.one/one_auth

# Enable dummy drivers
sudo chmod o+w /etc/one/oned.conf
sudo chmod o+w /etc/one/monitord.conf
echo 'IM_MAD = [ NAME="dummy", SUNSTONE_NAME="Dummy", EXECUTABLE="one_im_sh", ARGUMENTS="-r 3 -t 15 -w 90 dummy", THREADS=0]' >> /etc/one/monitord.conf
echo 'VM_MAD = [ NAME="dummy", SUNSTONE_NAME="Testing", EXECUTABLE="one_vmm_dummy",TYPE="xml" ]' >> /etc/one/oned.conf

# start oned
sudo systemctl start opennebula

# check it's up
timeout 60 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost 2633

# Create dummy host
onehost create dummy -i dummy -v dummy

