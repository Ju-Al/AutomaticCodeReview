 */
public class VirtualColumn implements DruidJson {
  private final String name;

  private final String expression;

  private final DruidType outputType;

  public VirtualColumn(String name, String expression, DruidType outputType) {
    this.name = Preconditions.checkNotNull(name);
    this.expression = Preconditions.checkNotNull(expression);
    this.outputType = outputType == null ? DruidType.FLOAT : outputType;
  }

  @Override public void write(JsonGenerator generator) throws IOException {
    generator.writeStartObject();
    generator.writeStringField("type", "expression");
    generator.writeStringField("name", name);
    generator.writeStringField("expression", expression);
    writeFieldIf(generator, "outputType", getOutputType().toString().toUpperCase(Locale.ENGLISH));
    generator.writeEndObject();
  }

  public String getName() {
    return name;
  }

  public String getExpression() {
    return expression;
  }

  public DruidType getOutputType() {
    return outputType;
  }

  /**
   * Virtual Column Builder
   */
  public static class Builder {
    private String name;

    private String expression;

    private DruidType type;

    public Builder withName(String name) {
      this.name = name;
      return this;
    }

    public Builder withExpression(String expression) {
      this.expression = expression;
      return this;
    }

    public Builder withType(DruidType type) {
      this.type = type;
      return this;
    }

    public VirtualColumn build() {
      return new VirtualColumn(name, expression, type);
    }
  }

  public static Builder builder() {
    return new Builder();
  }
}

// End VirtualColumn.java
