        $application->run(new ArrayInput([
            'command' => 'doctrine:database:drop',
            '--force' => '1',
        ]));

        $application->run(new ArrayInput([
            'command' => 'doctrine:database:create',
        ]));

        $command = new CreateSchemaDoctrineCommand();
        $application->add($command);
        $input = new ArrayInput([
            'command' => 'doctrine:schema:create',
        ]);
        $input->setInteractive(false);

        $command->run($input, new ConsoleOutput());

        $client = new Client($kernel);
        $client->request('GET', '/api/media');
        $this->assertSame(200, $client->getResponse()->getStatusCode());
    }
}
