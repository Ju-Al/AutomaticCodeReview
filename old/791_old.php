    {
        $expireDate = date('Y-m-d', time() - $daysOld * 24 * 60 * 60);
        $callback = function ($select) use ($expireDate, $idFrom, $idTo) {
            $where = $select->where->lessThan('created', $expireDate);
            if (null !== $idFrom) {
                $where->and->greaterThanOrEqualTo('id', $idFrom);
            }
            if (null !== $idTo) {
                $where->and->lessThanOrEqualTo('id', $idTo);
            }
        };
        return $this->delete($callback);
    }

    /**
     * Get the lowest id and highest id for expired sessions.
     *
     * @param int $daysOld Age in days of an "expired" session.
     *
     * @return array|bool Array of lowest id and highest id or false if no expired
     * records found
     */
    public function getExpiredIdRange($daysOld = 2)
    {
        $expireDate = date('Y-m-d', time() - $daysOld * 24 * 60 * 60);
        $callback = function ($select) use ($expireDate) {
            $select->where->lessThan('created', $expireDate);
        };
        $select = $this->getSql()->select();
        $select->columns(
            [
                'id' => new Expression('1'), // required for TableGateway
                'minId' => new Expression('MIN(id)'),
                'maxId' => new Expression('MAX(id)'),
            ]
        );
        $select->where($callback);
        $result = $this->selectWith($select)->current();
        if (null === $result->minId) {
            return false;
        }
        return [$result->minId, $result->maxId];
    }
}
