     */
    public function log($level, $message, array $context = [])
    {
        error_log($message);
    }
}
