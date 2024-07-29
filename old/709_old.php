        } catch (FileLocatorFileNotFoundException $e) {
            return [];
        }

        return array_map(function ($path) {
            return rtrim((new Filesystem())->makePathRelative($path, $this->projectDir), '/');
        }, $files);
    }

    private function executeRunonceFile(string $file): void
    {
        $this->framework->initialize();

        include $this->projectDir.'/'.$file;

        (new Filesystem())->remove($this->projectDir.'/'.$file);
    }

    private function executeSchemaDiff(bool $completeOption): bool
    {
        if (null === $this->installer) {
            $this->io->error('Service contao.installer of contao/installation-bundle not found.');

            return false;
        }

        $commandsByHash = [];

        while (true) {
            $this->installer->compileCommands();
            $commands = $this->installer->getCommands();

            if (!$commands) {
                return true;
            }

            $hasNewCommands = \count(
                array_filter(
                    array_keys(
                        array_merge(...array_values($commands))
                    ),
                    static function ($hash) use ($commandsByHash) {
                        return !isset($commandsByHash[$hash]);
                    }
                )
            );

            if (!$hasNewCommands) {
                return true;
            }

            $this->io->section('Pending database migrations');

            $commandsByHash = array_merge(...array_values($commands));

            $this->io->listing($commandsByHash);

            $options = ['yes', 'yes, with deletes', 'no'];

            if ($completeOption) {
                array_shift($options);
            }

            $answer = $this->io->choice(
                'Execute the listed database updates?',
                $options,
                $options[0]
            );

            if ('no' === $answer) {
                return false;
            }

            $this->io->section('Execute database migrations');

            $count = 0;

            foreach ($this->getCommandHashes($commands, 'yes, with deletes' === $answer) as $hash) {
                $this->io->writeln(' * '.$commandsByHash[$hash]);
                $this->installer->execCommand($hash);
                ++$count;
            }

            $this->io->success('Executed '.$count.' SQL queries.');
        }

        return true;
    }

    private function getCommandHashes(array $commands, bool $withDrops): array
    {
        if (!$withDrops) {
            unset($commands['ALTER_DROP']);

            foreach ($commands as $category => $commandsByHash) {
                foreach ($commandsByHash as $hash => $command) {
                    if ('DROP' === $category && false === strpos($command, 'DROP INDEX')) {
                        unset($commands[$category][$hash]);
                    }
                }
            }
        }

        return \count($commands) ? array_keys(array_merge(...array_values($commands))) : [];
    }
}
