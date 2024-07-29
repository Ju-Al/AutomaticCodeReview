        return $secure ? new SecureDelegator($handler) : $handler;
    }
}
