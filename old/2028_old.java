public final class BugCheckerAutoService extends BugChecker implements ClassTreeMatcher {

    private static final String AUTO_SERVICE = "com.google.auto.service.AutoService";
    private static final Matcher<ClassTree> isBugChecker =
            Matchers.allOf(Matchers.isSubtypeOf(BugChecker.class), Matchers.hasAnnotation(BugPattern.class));

    private static final Matcher<AnnotationTree> autoServiceBugChecker = Matchers.allOf(
            Matchers.isType(AUTO_SERVICE),
            Matchers.hasArgumentWithValue("value", Matchers.classLiteral(Matchers.isSameType(BugChecker.class))));

    @Override
    public Description matchClass(ClassTree classTree, VisitorState state) {
        if (!isBugChecker.matches(classTree, state)) {
            return Description.NO_MATCH;
        }

        TypeSymbol thisClassSymbol = ASTHelpers.getSymbol(classTree);
        if (thisClassSymbol.getKind() != ElementKind.CLASS) {
            return Description.NO_MATCH;
        }

        List<? extends AnnotationTree> annotations = ASTHelpers.getAnnotations(classTree);
        boolean hasAutoServiceBugChecker =
                annotations.stream().anyMatch(annotationTree -> autoServiceBugChecker.matches(annotationTree, state));
        if (hasAutoServiceBugChecker) {
            return Description.NO_MATCH;
        }

        SuggestedFix.Builder fix = SuggestedFix.builder();
        String autoService = SuggestedFixes.qualifyType(state, fix, AUTO_SERVICE);
        String bugChecker = SuggestedFixes.qualifyType(state, fix, BugChecker.class.getName());
        return buildDescription(classTree)
                .addFix(fix.prefixWith(classTree, "@" + autoService + "(" + bugChecker + ".class)\n")
                        .build())
                .build();
    }
}
