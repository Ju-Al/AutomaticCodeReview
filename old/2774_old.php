        try {
            $data = $this->getContainer()->get('wallabag_core.helper.entries_export')
                ->setEntries($entries)
                ->updateTitle('All')
                ->exportJsonData();
            file_put_contents($filePath, $data);
        } catch (\InvalidArgumentException $e) {
            $output->writeln(sprintf('<error>Error: "%s"</error>', $e->getMessage()));
        }

        $output->writeln('<info>Done.</info>');
    }

    /**
     * Fetches a user from its username.
     *
     * @param string $username
     *
     * @return \Wallabag\UserBundle\Entity\User
     */
    private function getUser($username)
    {
        return $this->getDoctrine()->getRepository('WallabagUserBundle:User')->findOneByUserName($username);
    }

    private function getDoctrine()
    {
        return $this->getContainer()->get('doctrine');
    }
}
