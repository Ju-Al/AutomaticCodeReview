
    public SearchedCase() {
        super(new Expression[4]);
    }

    public SearchedCase(Expression[] args) {
        super(args);
    }

    /**
     * Adds a new when clause.
     *
     * @param condition
     *            the condition
     * @param result
     *            the result for the specified condition
     */
    public void addWhen(Expression condition, Expression result) {
        int capacity = args.length;
        if (argsCount >= capacity) {
            args = Arrays.copyOf(args, capacity * 2);
        }
        args[argsCount++] = condition;
        args[argsCount++] = result;
    }

    /**
     * Adds an else clause.
     *
     * @param result
     *            the result
     */
    public void addElse(Expression result) {
        int capacity = args.length;
        if (argsCount >= capacity) {
            args = Arrays.copyOf(args, capacity * 2);
        }
        args[argsCount++] = result;
    }

    public void doneWithParameters() {
        if (args.length != argsCount) {
            args = Arrays.copyOf(args, argsCount);
        }
    }

    @Override
    public Value getValue(Session session) {
        int len = args.length - 1;
        for (int i = 0; i < len; i += 2) {
            if (args[i].getBooleanValue(session)) {
                return args[i + 1].getValue(session).convertTo(type, session);
            }
        }
        if ((len & 1) == 0) {
            return args[len].getValue(session).convertTo(type, session);
        }
        return ValueNull.INSTANCE;
    }

    @Override
    public Expression optimize(Session session) {
        TypeInfo typeInfo = TypeInfo.TYPE_UNKNOWN;
        int len = args.length - 1;
        boolean allConst = true;
        for (int i = 0; i < len; i += 2) {
            Expression condition = args[i].optimize(session);
            Expression result = args[i + 1].optimize(session);
            if (allConst) {
                if (condition.isConstant()) {
                    if (condition.getBooleanValue(session)) {
                        return result;
                    }
                } else {
                    allConst = false;
                }
            }
            args[i] = condition;
            args[i + 1] = result;
            typeInfo = SimpleCase.combineTypes(typeInfo, result);
        }
        if ((len & 1) == 0) {
            Expression result = args[len].optimize(session);
            if (allConst) {
                return result;
            }
            args[len] = result;
            typeInfo = SimpleCase.combineTypes(typeInfo, result);
        } else if (allConst) {
            return ValueExpression.NULL;
        }
        if (typeInfo.getValueType() == Value.UNKNOWN) {
            typeInfo = TypeInfo.TYPE_VARCHAR;
        }
        type = typeInfo;
        return this;
    }

    @Override
    public StringBuilder getSQL(StringBuilder builder, int sqlFlags) {
        builder.append("CASE");
        int len = args.length - 1;
        for (int i = 0; i < len; i += 2) {
            builder.append(" WHEN ");
            args[i].getSQL(builder, sqlFlags);
            builder.append(" THEN ");
            args[i + 1].getSQL(builder, sqlFlags);
        }
        if ((len & 1) == 0) {
            builder.append(" ELSE ");
            args[len].getSQL(builder, sqlFlags);
        }
        return builder.append(" END");
    }

}
