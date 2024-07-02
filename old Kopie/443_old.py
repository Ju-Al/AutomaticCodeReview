import asyncio
import json
import typing
from datetime import datetime
from datetime import datetime, timedelta

from app.service.base_service import BaseService


class AgentService(BaseService):

    def __init__(self):
    def __init__(self, untrusted_timer):
        self.log = self.add_service('agent_svc', self)
        self.loop = asyncio.get_event_loop()
        self.untrusted_timer = untrusted_timer

    async def start_sniffer_untrusted_agents(self):
        if self.untrusted_timer > 0:
            self.loop.create_task(self._sniffer_untrusted_agents())

    async def handle_heartbeat(self, paw, platform, server, group, executors, architecture, location, pid, ppid):
        self.log.debug('HEARTBEAT (%s)' % paw)
        agent = await self.get_service('data_svc').explode_agents(criteria=dict(paw=paw))
        now = self.get_current_timestamp()
        if agent:
            if agent[0]['trusted']:
                update_data = dict(last_seen=now, pid=pid, ppid=ppid, last_trusted_seen=now)
            else:
                update_data = dict(last_seen=now, pid=pid, ppid=ppid)
            await self.get_service('data_svc').update('core_agent', 'paw', paw, data=update_data)
            return agent[0]
        else:
            queued = dict(last_seen=now, paw=paw, platform=platform, server=server, host_group=group,
                          location=location, architecture=architecture, pid=pid, ppid=ppid,
                          trusted=True, last_trusted_seen=now)
            await self.get_service('data_svc').create_agent(agent=queued, executors=executors)
            return (await self.get_service('data_svc').explode_agents(criteria=dict(paw=paw)))[0]

    async def get_instructions(self, paw):
        commands = await self.get_service('data_svc').explode_chain(criteria=dict(paw=paw))
        instructions = []
        for link in [c for c in commands if not c['collect']]:
            await self.get_service('data_svc').update('core_chain', key='id', value=link['id'],
                                                      data=dict(collect=datetime.now()))
            payload = await self._gather_payload(link['ability'])
            instructions.append(json.dumps(dict(id=link['id'],
                                                sleep=link['jitter'],
                                                command=link['command'],
                                                executor=link['executor'],
                                                payload=payload)))
        return json.dumps(instructions)

    async def save_results(self, link_id, output, status):
        link = await self.get_service('data_svc').explode_chain(criteria=dict(id=link_id))
        op = (await self.get_service('data_svc').dao.get('core_operation', dict(id=link[0]['op_id'])))[0]
        agent = (await self.get_service('data_svc').dao.get('core_agent', dict(paw=link[0]['paw'])))[0]
        #last seen more accurate
        now = self.get_current_timestamp()
        if agent['trusted']:
            await self.get_service('data_svc').update('core_agent', 'paw', link[0]['paw'],
                                                    data=dict(last_seen=now, last_trusted_seen=now))
        else:
            await self.get_service('data_svc').update('core_agent', 'paw', link[0]['paw'],
                                                    data=dict(last_seen=now))
        #save result
        if (agent['trusted']) or (op['allow_untrusted']):
            await self.get_service('data_svc').create_result(result=dict(link_id=link_id, output=output))
            await self.get_service('data_svc').update('core_chain', key='id', value=link_id,
                                                    data=dict(status=int(status), finish=now))
        return json.dumps(dict(status=True))

    """ PRIVATE """

    async def _gather_payload(self, ability_id):
        payload = await self.get_service('data_svc').explode_payloads(criteria=dict(ability=ability_id))
        return payload[0]['payload'] if payload else ''

    async def perform_action(self, link: typing.Dict) -> int:
        """
        Perform a link in the context of an operation, respecting the 'run', 'paused' and 'run_one_step' operation
        states. Calling data_svc.create_link() directly will schedule the link for execution,
        ignoring the state of the operation.
        :param link: A link dictionary that has not yet been scheduled for execution using data_svc.create_link().
        :return: The id of the created link.
        """
        data_svc = self.get_service('data_svc')
        operation_svc = self.get_service('operation_svc')
        op_id = link['op_id']

        operation = (await data_svc.dao.get('core_operation', dict(id=op_id)))[0]
        while operation['state'] != operation_svc.op_states['RUNNING']:
            if operation['state'] == operation_svc.op_states['RUN_ONE_LINK']:
                link_id = await data_svc.create_link(link)
                await data_svc.dao.update('core_operation', 'id', op_id, dict(state=operation_svc.op_states['PAUSED']))
                return link_id
            else:
                await asyncio.sleep(30)
                operation = (await data_svc.dao.get('core_operation', dict(id=op_id)))[0]
        return await data_svc.create_link(link)

    async def _sniffer_untrusted_agents(self):
        data_svc = self.get_service('data_svc')
        next_check = self.untrusted_timer
        try:
            while True:
                await asyncio.sleep(next_check+1)
                trusted_agents = await data_svc.explode_agents(criteria=dict(trusted=1))
                now = datetime.now()
                next_check = self.untrusted_timer
                for a in trusted_agents:
                    last_trusted_seen = datetime.strptime(a['last_trusted_seen'], '%Y-%m-%d %H:%M:%S')
                    silence_time = (now-last_trusted_seen).total_seconds()
                    if silence_time > self.untrusted_timer:
                        await data_svc.update('core_agent', 'paw', a['paw'], data=dict(trusted=0))
                        self.log.debug('Agent %s is now untrusted' % a['paw'])
                    else:
                        trust_time_left = self.untrusted_timer - silence_time
                        if trust_time_left < next_check:
                            next_check = trust_time_left
        except Exception:
            traceback.print_exc()