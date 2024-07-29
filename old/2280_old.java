
        @Override
        public Object getNextToken() {
            if (iterator.hasNext()) {
                return new StringToken(iterator.next());
            } else {
                return null;
            }
        }

        @Override
        public void setFileName(final String fileName) {
        }
    }

    class DummyTokenFilter<T extends GenericToken> extends BaseTokenFilter<T> {

        Iterable<T> remainingTokens;

        DummyTokenFilter(final TokenManager tokenManager) {
            super(tokenManager);
        }

        @Override
        protected boolean shouldStopProcessing(final T currentToken) {
            return currentToken == null;
        }

        @Override
        protected void analyzeTokens(final T currentToken, final Iterable<T> remainingTokens) {
            this.remainingTokens = remainingTokens;
        }

        public Iterable getRemainingTokens() {
            return remainingTokens;
        }
    }

    @Test
    public void testRemainingTokensFunctionality1() {
        final TokenManager tokenManager = new StringTokenManager();
        final DummyTokenFilter tokenFilter = new DummyTokenFilter(tokenManager);
        final GenericToken firstToken = tokenFilter.getNextToken();
        assertEquals("a", firstToken.getImage());
        final Iterable<StringToken> iterable = tokenFilter.getRemainingTokens();
        final Iterator it1 = iterable.iterator();
        final Iterator it2 = iterable.iterator();
        assertTrue(it1.hasNext());
        assertTrue(it2.hasNext());
        final StringToken firstValFirstIt = (StringToken) it1.next();
        final StringToken firstValSecondIt = (StringToken) it2.next();
        assertTrue(it1.hasNext());
        assertTrue(it2.hasNext());
        final StringToken secondValFirstIt = (StringToken) it1.next();
        assertFalse(it1.hasNext());
        assertTrue(it2.hasNext());
        final StringToken secondValSecondIt = (StringToken) it2.next();
        assertFalse(it2.hasNext());
        assertEquals("b", firstValFirstIt.getImage());
        assertEquals("b", firstValSecondIt.getImage());
        assertEquals("c", secondValFirstIt.getImage());
        assertEquals("c", secondValSecondIt.getImage());
    }

    @Test
    public void testRemainingTokensFunctionality2() {
        final TokenManager tokenManager = new StringTokenManager();
        final DummyTokenFilter tokenFilter = new DummyTokenFilter(tokenManager);
        final GenericToken firstToken = tokenFilter.getNextToken();
        assertEquals("a", firstToken.getImage());
        final Iterable<StringToken> iterable = tokenFilter.getRemainingTokens();
        final Iterator it1 = iterable.iterator();
        final Iterator it2 = iterable.iterator();
        assertTrue(it1.hasNext());
        assertTrue(it2.hasNext());
        final StringToken firstValFirstIt = (StringToken) it1.next();
        assertTrue(it1.hasNext());
        final StringToken secondValFirstIt = (StringToken) it1.next();
        assertFalse(it1.hasNext());
        assertTrue(it2.hasNext());
        final StringToken firstValSecondIt = (StringToken) it2.next();
        assertTrue(it2.hasNext());
        final StringToken secondValSecondIt = (StringToken) it2.next();
        assertFalse(it2.hasNext());
        assertEquals("b", firstValFirstIt.getImage());
        assertEquals("b", firstValSecondIt.getImage());
        assertEquals("c", secondValFirstIt.getImage());
        assertEquals("c", secondValSecondIt.getImage());
    }

    @Test(expected = NoSuchElementException.class)
    public void testRemainingTokensFunctionality3() {
        final TokenManager tokenManager = new StringTokenManager();
        final DummyTokenFilter tokenFilter = new DummyTokenFilter(tokenManager);
        final GenericToken firstToken = tokenFilter.getNextToken();
        assertEquals("a", firstToken.getImage());
        final Iterable<StringToken> iterable = tokenFilter.getRemainingTokens();
        final Iterator it1 = iterable.iterator();
        final Iterator it2 = iterable.iterator();
        it1.next();
        it1.next();
        it2.next();
        it2.next();
        it1.next();
    }

    @Test(expected = ConcurrentModificationException.class)
    public void testRemainingTokensFunctionality4() {
        final TokenManager tokenManager = new StringTokenManager();
        final DummyTokenFilter tokenFilter = new DummyTokenFilter(tokenManager);
        final GenericToken firstToken = tokenFilter.getNextToken();
        assertEquals("a", firstToken.getImage());
        final Iterable<StringToken> iterable = tokenFilter.getRemainingTokens();
        final Iterator it1 = iterable.iterator();
        final GenericToken secondToken = tokenFilter.getNextToken();
        assertEquals("b", secondToken.getImage());
        it1.next();
    }

}
