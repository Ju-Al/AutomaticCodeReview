
        /**
         * Returns the Visibility enum key for a node
         *
         * @param node A node
         *
         * @return The visibility enum key for a node
         */
        public static Visibility get(AbstractJavaAccessNode node) {
            return node.isPublic() ? PUBLIC
                    : node.isPackagePrivate() ? PACKAGE
                    : node.isProtected() ? PROTECTED
                    : node.isPrivate() ? PRIVATE : UNDEF;
        }
    }


}
