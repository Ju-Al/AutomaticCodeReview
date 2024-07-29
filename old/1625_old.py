
    def test_run(self):
        message = "test message"
        sch = luigi.scheduler.CentralPlannerScheduler()
        with luigi.worker.Worker(scheduler=sch) as w:
            class MyTask(luigi.Task):
                def run(self, status_message_callback):
                    status_message_callback(message)

            task = MyTask()
            w.add(task)
            w.run()

            self.assertEqual(sch.get_task_status_message(task.task_id)["statusMessage"], message)
