            JsonNode jsonNode = jsonObjectMapper.readTree(jsonString);
            YAMLFactory yamlFactory = new YAMLFactory()
                    .disable(YAMLGenerator.Feature.MINIMIZE_QUOTES)
                    .disable(YAMLGenerator.Feature.WRITE_DOC_START_MARKER);
            return new YAMLMapper(yamlFactory).writeValueAsString(jsonNode);
        } catch (Exception e) {
            logger.error("Error mapping to yaml: {}", e.getMessage());
            return null;
        }
    }
}
