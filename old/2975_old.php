        }

        return $this->headerBag;
    }

    public function mapToResponse(Response $response): void
    {
        foreach ($this->getHeaderBag()->all() as $name => $values) {
            $response->headers->set($name, $values, false); // Do not replace but add
        }
    }
}
