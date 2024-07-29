    {
      for (vector<long>::const_iterator nodeIt = nodes.begin(); nodeIt != nodes.end(); ++nodeIt)
      {
        ConstNodePtr node = _map->getNode(*nodeIt);
        if (node.get() == NULL)
        {
          valid = false;
          break;
        }
      }
    }
    //  Write out the way in geojson if valid
    if (valid)
      _writeWay(w);
    else
    {
      for (vector<long>::const_iterator nodeIt = nodes.begin(); nodeIt != nodes.end(); ++nodeIt)
      {
        ConstNodePtr node = _map->getNode(*nodeIt);
        if (node.get() != NULL)
        {
          if (_firstElement)
            _firstElement = false;
          else
            _write(",", true);
          _writeNode(node);
        }
      }
    }
  }
}

void OsmGeoJsonWriter::_writeWay(ConstWayPtr way)
{
  if (way.get() == NULL)
    return;
  //  Write out the way in geojson
  _write("{");
  _writeFeature(way);
  _write(",");
  _write("\"geometry\": {");
  _writeGeometry(way);
  _write("}");
  _write("}", false);

}

void OsmGeoJsonWriter::_writeRelations()
{
  const RelationMap& relations = _map->getRelations();
  for (RelationMap::const_iterator it = relations.begin(); it != relations.end(); ++it)
  {
    if (_firstElement)
      _firstElement = false;
    else
      _write(",", true);
    ConstRelationPtr r = it->second;
    //  Write out the relation (and all children) in geojson
    _write("{");
    _writeFeature(r);
    _write(",");
    _write("\"geometry\": {");
    _writeGeometry(r);
    _write("}");
    _write("}", false);
  }
}

void OsmGeoJsonWriter::_writeRelationInfo(ConstRelationPtr r)
{
  _writeKvp("relation-type", r->getType()); _write(",");
  QString roles = _buildRoles(r).c_str();
  _writeKvp("roles", roles);
}

string OsmGeoJsonWriter::_buildRoles(ConstRelationPtr r)
{
  bool first = true;
  return _buildRoles(r, first);
}

string OsmGeoJsonWriter::_buildRoles(ConstRelationPtr r, bool& first)
{
  stringstream ss;
  const vector<RelationData::Entry>& members = r->getMembers();
  //  Iterate all members and concatenate the roles separated by a semicolon
  for (vector<RelationData::Entry>::const_iterator it = members.begin(); it != members.end(); ++it)
  {
    ConstElementPtr e = _map->getElement(it->getElementId());
    if (e.get() == NULL)
      continue;
    if (first)
      first = false;
    else
      ss << ";";
    ss << it->getRole();
    //  Recursively get the roles of the sub-relation that is valid
    if (it->getElementId().getType() == ElementType::Relation &&
        _map->getRelation(it->getElementId().getId()).get() != NULL)
    {
      ss << ";" << _buildRoles(_map->getRelation(it->getElementId().getId()), first);
    }
  }
  return ss.str();
}

}
