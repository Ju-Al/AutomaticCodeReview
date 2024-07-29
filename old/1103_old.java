          throw e;
        }
        return TypeInfoFactory.getDecimalTypeInfo(precision, scale);
      case STRUCT:
        return generateStructTypeInfo((Types.StructType) type);
      case LIST:
        return generateListTypeInfo((Types.ListType) type);
      case MAP:
        return generateMapTypeInfo((Types.MapType) type);
      default:
        throw new SerDeException("Can't map Iceberg type to Hive TypeInfo: '" + type.typeId() + "'");
    }
  }

  private static TypeInfo generateMapTypeInfo(Types.MapType type) throws Exception {
    Type keyType = type.keyType();
    Type valueType = type.valueType();
    return TypeInfoFactory.getMapTypeInfo(generateTypeInfo(keyType), generateTypeInfo(valueType));
  }

  private static TypeInfo generateStructTypeInfo(Types.StructType type) throws Exception {
    List<Types.NestedField> fields = type.fields();
    List<String> fieldNames = new ArrayList<>(fields.size());
    List<TypeInfo> typeInfos = new ArrayList<>(fields.size());

    for (Types.NestedField field : fields) {
      fieldNames.add(field.name());
      typeInfos.add(generateTypeInfo(field.type()));
    }
    return TypeInfoFactory.getStructTypeInfo(fieldNames, typeInfos);
  }

  private static TypeInfo generateListTypeInfo(Types.ListType type) throws Exception {
    return TypeInfoFactory.getListTypeInfo(generateTypeInfo(type.elementType()));
  }
}
