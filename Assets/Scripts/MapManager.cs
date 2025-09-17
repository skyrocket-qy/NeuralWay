using UnityEngine;
using System.Collections.Generic;

public class MapManager : MonoBehaviour
{
    public List<MapData> availableMaps = new List<MapData>();
    public int selectedMapIndex = 0;
    private MapData _currentMapData; // Renamed to avoid confusion with the property

    public GameObject startMarkerPrefab; // Assign a prefab for the start point visual
    public GameObject endMarkerPrefab;   // Assign a prefab for the end point visual
    public GameObject barrierPrefab;     // Assign a prefab for barriers
    public GameObject placementSpotPrefab; // Assign a prefab for hero placement spots

    private GameObject _startMarkerInstance;
    private GameObject _endMarkerInstance;
    private List<GameObject> _barrierInstances = new List<GameObject>();
    private List<GameObject> _placementSpotInstances = new List<GameObject>();
    private GameObject _pathInstance;

    public Vector3 MonsterStartPoint => _currentMapData != null ? _currentMapData.monsterStartPoint : Vector3.zero;
    public Vector3 MonsterEndPoint => _currentMapData != null ? _currentMapData.monsterEndPoint : Vector3.zero;

    void Awake()
    {
        LoadSelectedMap();
    }

    public void LoadSelectedMap()
    {
        if (availableMaps == null || availableMaps.Count == 0)
        {
            Debug.LogError("No maps available in MapManager! Please assign MapData assets to the list.");
            return;
        }

        if (selectedMapIndex < 0 || selectedMapIndex >= availableMaps.Count)
        {
            Debug.LogWarning($"Selected map index {selectedMapIndex} is out of bounds. Loading first map instead.");
            selectedMapIndex = 0;
        }

        LoadMap(availableMaps[selectedMapIndex]);
    }

    public void LoadMap(MapData mapData)
    {
        ClearCurrentMap();
        _currentMapData = mapData; // Assign to the private field

        if (_currentMapData == null)
        {
            Debug.LogError("Attempted to load null MapData.");
            return;
        }

        // Instantiate Start Marker
        if (startMarkerPrefab != null)
        {
            _startMarkerInstance = Instantiate(startMarkerPrefab, _currentMapData.monsterStartPoint, Quaternion.identity);
            _startMarkerInstance.name = "Start Marker";
        }
        else
        {
            Debug.LogWarning("Start Marker Prefab not assigned. Creating a default sphere.");
            _startMarkerInstance = GameObject.CreatePrimitive(PrimitiveType.Sphere);
            _startMarkerInstance.transform.position = _currentMapData.monsterStartPoint;
            _startMarkerInstance.name = "Start Marker (Default)";
        }

        // Instantiate End Marker
        if (endMarkerPrefab != null)
        {
            _endMarkerInstance = Instantiate(endMarkerPrefab, _currentMapData.monsterEndPoint, Quaternion.identity);
            _endMarkerInstance.name = "End Marker";
        }
        else
        {
            Debug.LogWarning("End Marker Prefab not assigned. Creating a default cube.");
            _endMarkerInstance = GameObject.CreatePrimitive(PrimitiveType.Cube);
            _endMarkerInstance.transform.position = _currentMapData.monsterEndPoint;
            _endMarkerInstance.name = "End Marker (Default)";
        }

        // Determine which barrier prefab to use
        GameObject currentBarrierPrefab = _currentMapData.mapSpecificBarrierPrefab != null ? _currentMapData.mapSpecificBarrierPrefab : barrierPrefab;
        if (currentBarrierPrefab != null)
        {
            foreach (Vector3 pos in _currentMapData.barrierPositions)
            {
                GameObject barrier = Instantiate(currentBarrierPrefab, pos, Quaternion.identity);
                barrier.name = "Barrier";
                _barrierInstances.Add(barrier);
            }
        }
        else if (_currentMapData.barrierPositions.Count > 0)
        {
            Debug.LogWarning("Barrier Prefab not assigned in MapManager or MapData. Barriers will not be instantiated.");
        }

        // Determine which placement spot prefab to use
        GameObject currentPlacementSpotPrefab = _currentMapData.mapSpecificPlacementSpotPrefab != null ? _currentMapData.mapSpecificPlacementSpotPrefab : placementSpotPrefab;
        if (currentPlacementSpotPrefab != null)
        {
            foreach (Vector3 pos in _currentMapData.heroPlacementSpots)
            {
                GameObject spot = Instantiate(currentPlacementSpotPrefab, pos, Quaternion.identity);
                spot.name = "Placement Spot";
                _placementSpotInstances.Add(spot);
            }
        }
        else if (_currentMapData.heroPlacementSpots.Count > 0)
        {
            Debug.LogWarning("Placement Spot Prefab not assigned in MapManager or MapData. Placement spots will not be instantiated.");
        }

        // Instantiate Path Visual (if specified)
        GameObject currentPathPrefab = _currentMapData.mapSpecificPathPrefab;
        if (currentPathPrefab != null)
        {
            Vector3 start = _currentMapData.monsterStartPoint;
            Vector3 end = _currentMapData.monsterEndPoint;

            // Calculate direction and distance
            Vector3 direction = (end - start).normalized;
            float distance = Vector3.Distance(start, end);

            // Instantiate the path prefab at the start point
            _pathInstance = Instantiate(currentPathPrefab, start, Quaternion.identity);
            _pathInstance.name = "Monster Path Visual";

            // Position the path prefab correctly
            // Assuming the prefab's forward (Z-axis) points along the path
            _pathInstance.transform.position = start + (end - start) / 2f; // Center the path
            _pathInstance.transform.rotation = Quaternion.LookRotation(direction); // Rotate to face end point
            _pathInstance.transform.localScale = new Vector3(_pathInstance.transform.localScale.x, _pathInstance.transform.localScale.y, distance); // Stretch along Z

            // If it's a 2D game, you might need to adjust the rotation for 2D sprites
            // For 2D, you might want to rotate around the Z-axis
            // float angle = Mathf.Atan2(direction.y, direction.x) * Mathf.Rad2Deg;
            // _pathInstance.transform.rotation = Quaternion.Euler(0, 0, angle);
            // _pathInstance.transform.localScale = new Vector3(distance, _pathInstance.transform.localScale.y, _pathInstance.transform.localScale.z);
        }

        Debug.Log($"Map '{_currentMapData.mapName}' loaded successfully.");
    }

    public void SelectAndLoadMap(int index)
    {
        if (index >= 0 && index < availableMaps.Count)
        {
            selectedMapIndex = index;
            LoadMap(availableMaps[selectedMapIndex]);
        }
        else
        {
            Debug.LogWarning($"Attempted to select map with invalid index: {index}. Available maps: {availableMaps.Count}");
        }
    }

    public void SelectAndLoadMap(string mapName)
    {
        for (int i = 0; i < availableMaps.Count; i++)
        {
            if (availableMaps[i].mapName == mapName)
            {
                selectedMapIndex = i;
                LoadMap(availableMaps[selectedMapIndex]);
                return;
            }
        }
        Debug.LogWarning($"Attempted to select map with unknown name: {mapName}");
    }

    private void ClearCurrentMap()
    {
        if (_startMarkerInstance != null) Destroy(_startMarkerInstance);
        if (_endMarkerInstance != null) Destroy(_endMarkerInstance);
        foreach (GameObject barrier in _barrierInstances) Destroy(barrier);
        foreach (GameObject spot in _placementSpotInstances) Destroy(spot);

        _barrierInstances.Clear();
        _placementSpotInstances.Clear();

        // Clear path instances if any
        if (_pathInstance != null) Destroy(_pathInstance);
    }
}
