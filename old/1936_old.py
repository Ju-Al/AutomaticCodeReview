# Copyright 2018 Microsoft Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Requires Python 2.6+ and Openssl 1.0+
#

import re

from azurelinuxagent.common.utils.textutil import parse_doc, find, findall
from tests.tools import load_bin_data, load_data, MagicMock, Mock
from azurelinuxagent.common.exception import HttpError, ResourceGoneError
from azurelinuxagent.common.future import httpclient
from azurelinuxagent.common.utils.cryptutil import CryptUtil

DATA_FILE = {
        "version_info": "wire/version_info.xml",
        "goal_state": "wire/goal_state.xml",
        "hosting_env": "wire/hosting_env.xml",
        "shared_config": "wire/shared_config.xml",
        "certs": "wire/certs.xml",
        "ext_conf": "wire/ext_conf.xml",
        "manifest": "wire/manifest.xml",
        "ga_manifest": "wire/ga_manifest.xml",
        "trans_prv": "wire/trans_prv",
        "trans_cert": "wire/trans_cert",
        "test_ext": "ext/sample_ext-1.3.0.zip",
        "remote_access": None,
        "in_vm_artifacts_profile": None
}
from tests.protocol.mockwiredata_fetcher import DEFAULT_FETCHER, data_files_to_fetcher, generate_ext_fetcher_func

def get_file_based_wire_protocol_data(data_files):
    return WireProtocolData(data_files)

def get_dynamic_wire_protocol_data(data_fetcher):
    return WireProtocolDataFromMemory(data_fetcher)

class WireProtocolDataBase(object):
    def __init__(self):
        self.emulate_stale_goal_state = False
        self.call_counts = {
            "comp=versions": 0,
            "/versions": 0,
            "/health": 0,
            "/HealthService": 0,
            "goalstate": 0,
            "hostingenvuri": 0,
            "sharedconfiguri": 0,
            "certificatesuri": 0,
            "extensionsconfiguri": 0,
            "remoteaccessinfouri": 0,
            "extensionArtifact": 0,
            "manifest.xml": 0,
            "manifest_of_ga.xml": 0,
            "ExampleHandlerLinux": 0,
            "in_vm_artifacts_profile": 0
        }
        self.version_info = None
        self.goal_state = None
        self.hosting_env = None
        self.shared_config = None
        self.certs = None
        self.ext_conf = None
        self.manifest = None
        self.ga_manifest = None
        self.trans_prv = None
        self.trans_cert = None
        self.ext = None
        self.remote_access = None
        self.in_vm_artifacts_profile = None


    def reload(self):
        raise  NotImplementedError("WireProtocolDataBase by itself has no data with which to load.")

    def mock_http_get(self, url, *args, **kwargs):
        content = None

        resp = MagicMock()
        resp.status = httpclient.OK

        if "comp=versions" in url:  # wire server versions
            content = self.version_info
            self.call_counts["comp=versions"] += 1
        elif "/versions" in url:  # HostPlugin versions
            content = '["2015-09-01"]'
            self.call_counts["/versions"] += 1
        elif url.endswith("/health"):  # HostPlugin health
            content = ''
            self.call_counts["/health"] += 1
        elif "goalstate" in url:
            content = self.goal_state
            self.call_counts["goalstate"] += 1
        elif "hostingenvuri" in url:
            content = self.hosting_env
            self.call_counts["hostingenvuri"] += 1
        elif "sharedconfiguri" in url:
            content = self.shared_config
            self.call_counts["sharedconfiguri"] += 1
        elif "certificatesuri" in url:
            content = self.certs
            self.call_counts["certificatesuri"] += 1
        elif "extensionsconfiguri" in url:
            content = self.ext_conf
            self.call_counts["extensionsconfiguri"] += 1
        elif "remoteaccessinfouri" in url:
            content = self.remote_access
            self.call_counts["remoteaccessinfouri"] += 1
        elif ".vmSettings" in url or ".settings" in url:
            content = self.in_vm_artifacts_profile
            self.call_counts["in_vm_artifacts_profile"] += 1

        else:
            # A stale GoalState results in a 400 from the HostPlugin
            # for which the HTTP handler in restutil raises ResourceGoneError
            if self.emulate_stale_goal_state:
                if "extensionArtifact" in url:
                    self.emulate_stale_goal_state = False
                    self.call_counts["extensionArtifact"] += 1
                    raise ResourceGoneError()
                else:
                    raise HttpError()

            # For HostPlugin requests, replace the URL with that passed
            # via the x-ms-artifact-location header
            if "extensionArtifact" in url:
                self.call_counts["extensionArtifact"] += 1
                if "headers" not in kwargs:
                    raise ValueError("HostPlugin request is missing the HTTP headers: {0}", kwargs)
                if "x-ms-artifact-location" not in kwargs["headers"]:
                    raise ValueError("HostPlugin request is missing the x-ms-artifact-location header: {0}", kwargs)
                url = kwargs["headers"]["x-ms-artifact-location"]

            if "manifest.xml" in url:
                content = self.manifest
                self.call_counts["manifest.xml"] += 1
            elif "manifest_of_ga.xml" in url:
                content = self.ga_manifest
                self.call_counts["manifest_of_ga.xml"] += 1
            elif "ExampleHandlerLinux" in url:
                content = self.ext
                self.call_counts["ExampleHandlerLinux"] += 1
                resp.read = Mock(return_value=content)
                return resp
            elif ".vmSettings" in url or ".settings" in url:
                content = self.in_vm_artifacts_profile
                self.call_counts["in_vm_artifacts_profile"] += 1
            else:
                raise Exception("Bad url {0}".format(url))

        resp.read = Mock(return_value=content.encode("utf-8"))
        return resp

    def mock_http_post(self, url, *args, **kwargs):
        content = None

        resp = MagicMock()
        resp.status = httpclient.OK

        if url.endswith('/HealthService'):
            self.call_counts['/HealthService'] += 1
            content = ''
        else:
            raise Exception("Bad url {0}".format(url))

        resp.read = Mock(return_value=content.encode("utf-8"))
        return resp

    def mock_crypt_util(self, *args, **kw):
        #Partially patch instance method of class CryptUtil
        cryptutil = CryptUtil(*args, **kw)
        cryptutil.gen_transport_cert = Mock(side_effect=self.mock_gen_trans_cert)
        return cryptutil
    
    def mock_gen_trans_cert(self, trans_prv_file, trans_cert_file):
        with open(trans_prv_file, 'w+') as prv_file:
            prv_file.write(self.trans_prv)

        with open(trans_cert_file, 'w+') as cert_file:
            cert_file.write(self.trans_cert)

    def get_no_of_plugins_in_extension_config(self):
        if self.ext_conf is None:
            return 0
        ext_config_doc = parse_doc(self.ext_conf)
        plugins_list = find(ext_config_doc, "Plugins")
        return len(findall(plugins_list, "Plugin"))

    #
    # Having trouble reading the regular expressions below? you are not alone!
    #
    # For the use of "(?<=" "(?=" see 7.2.1 in https://docs.python.org/3.1/library/re.html
    # For the use of "\g<1>" see backreferences in https://docs.python.org/3.1/library/re.html#re.sub
    #
    # Note that these regular expressions are not enough to parse all valid XML documents (e.g. they do
    # not account for metacharacters like < or > in the values) but they are good enough for the test
    # data. There are some basic checks, but the functions may not match valid XML or produce invalid
    # XML if their input is too complex.
    #
    @staticmethod
    def replace_xml_element_value(xml_document, element_name, element_value):
        new_xml_document = re.sub(r'(?<=<{0}>).+(?=</{0}>)'.format(element_name), element_value, xml_document)
        if new_xml_document == xml_document:
            raise Exception("Could not match element '{0}'", element_name)
        return new_xml_document

    @staticmethod
    def replace_xml_attribute_value(xml_document, element_name, attribute_name, attribute_value):
        new_xml_document = re.sub(r'(?<=<{0} )(.*{1}=")[^"]+(?="[^>]*>)'.format(element_name, attribute_name), r'\g<1>{0}'.format(attribute_value), xml_document)
        if new_xml_document == xml_document:
            raise Exception("Could not match attribute '{0}' of element '{1}'".format(attribute_name, element_name))
        return new_xml_document

    @staticmethod
    def replace_xml_attribute_value_on_matching_element(xml_document, element_name, identifying_attr_str, attribute_name, attribute_value):

        # Two cases: the attribute we're using to identify the element might be before the attribute we're replacing,
        # or it might be after. We need to handle both.
        replacing_attr_after_identifying_attr = r'(?<=<{0} )(.*{1}.*{2}=")[^"]+(?="[^>]*>)'.format(element_name, identifying_attr_str, attribute_name)
        replacing_attr_before_identifying_attr = r'(?<=<{0} )(.*{1}=")[^"]+(?=".*{2}[^>]*>)'.format(element_name, attribute_name, identifying_attr_str)

        for regex in [replacing_attr_before_identifying_attr, replacing_attr_after_identifying_attr]:
            new_xml = re.sub(regex, r'\g<1>{0}'.format(attribute_value), xml_document)
            if new_xml != xml_document:
                return new_xml
        
        raise Exception("Could not match attribute '{0}' of element '{1}' with '{2}'".format(attribute_name, element_name, identifying_attr_str))

    def set_incarnation(self, incarnation):
        '''
        Sets the incarnation in the goal state, but not on its subcomponents (e.g. hosting env, shared config)
        '''
        self.goal_state = WireProtocolDataBase.replace_xml_element_value(self.goal_state, "Incarnation", str(incarnation))

    def set_container_id(self, container_id):
        self.goal_state = WireProtocolDataBase.replace_xml_element_value(self.goal_state, "ContainerId", container_id)

    def set_role_config_name(self, role_config_name):
        self.goal_state = WireProtocolDataBase.replace_xml_element_value(self.goal_state, "ConfigName", role_config_name)

    def set_hosting_env_deployment_name(self, deployment_name):
        self.hosting_env = WireProtocolDataBase.replace_xml_attribute_value(self.hosting_env, "Deployment", "name", deployment_name)

    def set_shared_config_deployment_name(self, deployment_name):
        self.shared_config = WireProtocolDataBase.replace_xml_attribute_value(self.shared_config, "Deployment", "name", deployment_name)

    def set_extensions_config_sequence_number(self, sequence_number):
        '''
        Sets the sequence number for *all* extensions
        '''
        self.ext_conf = WireProtocolDataBase.replace_xml_attribute_value(self.ext_conf, "RuntimeSettings", "seqNo", str(sequence_number))

    def set_extensions_config_version(self, version):
        '''
        Sets the version for *all* extensions
        '''
        self.ext_conf = WireProtocolDataBase.replace_xml_attribute_value(self.ext_conf, "Plugin", "version", version)
    
    def set_specific_extension_config_version(self, extensionName, version):
        """
        Set the version for the only the specified extension.
        """
        self.ext_conf = WireProtocolDataBase.replace_xml_attribute_value_on_matching_element(self.ext_conf, "Plugin", 'name="{0}"'.format(extensionName), "version", version)

    def set_extensions_config_state(self, state):
        '''
        Sets the state for *all* extensions
        '''
        self.ext_conf = WireProtocolDataBase.replace_xml_attribute_value(self.ext_conf, "Plugin", "state", state)

    def set_manifest_version(self, version):
        '''
        Sets the version of the extension manifest
        '''
        self.manifest = WireProtocolDataBase.replace_xml_element_value(self.manifest, "Version", version)

class WireProtocolData(WireProtocolDataBase):

    def __init__(self, data_files=DATA_FILE):
        super(WireProtocolData, self).__init__()
        self.data_files = data_files
        self.reload()

    def reload(self):
        """
        Loads into this object's member variables the file contents given by self.data_files.
        """
        self.version_info = load_data(self.data_files.get("version_info"))
        self.goal_state = load_data(self.data_files.get("goal_state"))
        self.hosting_env = load_data(self.data_files.get("hosting_env"))
        self.shared_config = load_data(self.data_files.get("shared_config"))
        self.certs = load_data(self.data_files.get("certs"))
        self.ext_conf = self.data_files.get("ext_conf")
        if self.ext_conf is not None:
            self.ext_conf = load_data(self.ext_conf)
        self.manifest = load_data(self.data_files.get("manifest"))
        self.ga_manifest = load_data(self.data_files.get("ga_manifest"))
        self.trans_prv = load_data(self.data_files.get("trans_prv"))
        self.trans_cert = load_data(self.data_files.get("trans_cert"))
        self.ext = load_bin_data(self.data_files.get("test_ext"))

        remote_access_data_file = self.data_files.get("remote_access")
        if remote_access_data_file is not None:
            self.remote_access = load_data(remote_access_data_file)

        in_vm_artifacts_profile_file = self.data_files.get("in_vm_artifacts_profile")
        if in_vm_artifacts_profile_file is not None:
            self.in_vm_artifacts_profile = load_data(in_vm_artifacts_profile_file)


class WireProtocolDataFromMemory(WireProtocolDataBase):

    def __init__(self, data_fetcher=DEFAULT_FETCHER):
        super(WireProtocolDataFromMemory, self).__init__()
        self.data_fetcher = data_fetcher
        self.reload()
    
    def reload(self):
        """
        Loads into this object's member variables the dynamically generated contents given
        by self.data_fetcher.
        """
        self.version_info = self.data_fetcher['version_info']()
        self.goal_state = self.data_fetcher['goal_state']()
        self.hosting_env = self.data_fetcher['hosting_env']()
        self.shared_config = self.data_fetcher['shared_config']()
        self.certs = self.data_fetcher['certs']()
        self.ext_conf = self.data_fetcher['ext_conf']()
        self.manifest = self.data_fetcher['manifest']()
        self.ga_manifest = self.data_fetcher['ga_manifest']()
        self.trans_prv = self.data_fetcher['trans_prv']()
        self.trans_cert = self.data_fetcher['trans_cert']()
        self.ext = self.data_fetcher['test_ext']()
        self.remote_access = self.data_fetcher['remote_access']()
        self.in_vm_artifacts_profile = self.data_fetcher['in_vm_artifacts_profile']()