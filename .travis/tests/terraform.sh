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

# define env variables
export OPENNEBULA_ENDPOINT="http://localhost:2633/RPC2"
export OPENNEBULA_USERNAME="oneadmin"
export OPENNEBULA_PASSWORD="opennebula"
export TF_ACC=1
export CODECOV_TOKEN="b4d4b95c-bb63-4412-b3aa-6d0cf674d619"

# install dep dependencies manager
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# get dependencies
dep ensure

# build addon
go build -o terraform-provider-opennebula

# Run tests
echo "" > coverage.txt
cd opennebula && go test -coverprofile=profile.out -v
if [ -f profile.out ]; then
    cat profile.out >> coverage.txt
    go tool cover -func=profile.out
    rm profile.out
fi

