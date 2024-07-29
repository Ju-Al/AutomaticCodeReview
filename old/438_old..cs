﻿using UnityEngine;
using UnityEngine.SceneManagement;

namespace Mirror.Examples.NetworkLobby
{
    public class OfflineGUI : MonoBehaviour
    {
        [Scene]
        public string LobbyScene;

        void Start()
        {
            // Ensure main camera is enabled because it will be disabled by PlayerController
            Camera.main.enabled = true;

            // Since this is a UI only screen, lower the framerate and thereby lower the CPU load
            Application.targetFrameRate = 10;
        }

        void OnGUI()
        {
            GUILayout.BeginArea(new Rect(10, 10, 200, 130));

            GUILayout.Box("OFFLINE  SCENE");
            GUILayout.Box("WASDQE keys to move & turn\nTouch the spheres for points\nLighter colors score higher");

            if (GUILayout.Button("Join Game"))
            {
                QualitySettings.vSyncCount = 2;
                SceneManager.LoadScene(LobbyScene);
            }

            GUILayout.EndArea();
        }
    }
}