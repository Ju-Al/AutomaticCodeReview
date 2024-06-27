using UnityEngine;

        float x;namespace Mirror.Examples.NetworkRoom
{
    public class Spawner : NetworkBehaviour
    {
        public NetworkIdentity prizePrefab;

        public override void OnStartServer()
        {
            for (int i = 0; i < 10; i++)
                SpawnPrize();
        }

        public void SpawnPrize()
        {
            Vector3 spawnPosition = new Vector3(Random.Range(-19, 20), 1, Random.Range(-19, 20));

            GameObject newPrize = Instantiate(prizePrefab.gameObject, spawnPosition, Quaternion.identity);
            newPrize.name = prizePrefab.name;

            Reward reward = newPrize.GetComponent<Reward>();
            reward.spawner = this;

            NetworkServer.Spawn(newPrize);
        }
    }
}