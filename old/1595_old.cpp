 *                                                                            *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may    *
 * not use this file except in compliance with the License. You may obtain a  *
 * copy of the License at http://www.apache.org/licenses/LICENSE-2.0.         *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 * -------------------------------------------------------------------------- */

//=============================================================================
// INCLUDES
//=============================================================================
#include "AbstractPathPoint.h"
#include "OpenSim/Simulation/Model/PhysicalFrame.h"

//=============================================================================
// STATICS
//=============================================================================
using namespace std;
using namespace OpenSim;
using SimTK::Vec3;
using SimTK::Transform;


const PhysicalFrame& AbstractPathPoint::getParentFrame() const
{
    return getSocket<PhysicalFrame>("parent_frame").getConnectee();
}

void AbstractPathPoint::setParentFrame(const OpenSim::PhysicalFrame& frame)
{
    connectSocket_parent_frame(frame);
}

const PhysicalFrame& AbstractPathPoint::getBody() const {
    return getParentFrame();
}

void AbstractPathPoint::setBody(const PhysicalFrame& body)
{
    setParentFrame(body);
}

const std::string& AbstractPathPoint::getBodyName() const {
    return getParentFrame().getName();
}

void AbstractPathPoint::updateFromXMLNode(SimTK::Xml::Element& aNode, int versionNumber)
{
    int documentVersion = versionNumber;
    if (documentVersion < XMLDocument::getLatestVersion()) {
        if (documentVersion < 30505) {
            // replace old properties with latest use of Connectors
            SimTK::Xml::element_iterator bodyElement = aNode.element_begin("body");
            std::string bodyName("");
            if (bodyElement != aNode.element_end()) {
                bodyElement->getValueAs<std::string>(bodyName);
                XMLDocument::addConnector(aNode, "Connector_PhysicalFrame_",
                    "parent_frame", bodyName);
            }
        }
    }

    Super::updateFromXMLNode(aNode, versionNumber);
}

