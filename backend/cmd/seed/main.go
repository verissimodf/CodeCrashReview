package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"codecrashreview/backend/internal/config"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Seed challenge templates
	if err := seedChallengeTemplates(ctx, pool); err != nil {
		log.Fatalf("Failed to seed challenge templates: %v", err)
	}

	// Seed some example snippets
	if err := seedSnippets(ctx, pool); err != nil {
		log.Fatalf("Failed to seed snippets: %v", err)
	}

	log.Println("Database seeded successfully")
}

func seedChallengeTemplates(ctx context.Context, pool *pgxpool.Pool) error {
	templates := []map[string]interface{}{
		{
			"name":                       "security-fix",
			"description":                "Template for security vulnerability fixes",
			"category":                   "security",
			"default_severity":           "high",
			"default_time_limit_seconds": 300,
			"default_max_turns":          10,
			"acceptance_criteria_template": []map[string]interface{}{
				{"id": "ac1", "description": "No security vulnerabilities detected by static analysis", "type": "static_analysis", "weight": 0.4, "pass_threshold": 1.0},
				{"id": "ac2", "description": "All existing tests pass", "type": "test", "weight": 0.3, "pass_threshold": 1.0},
				{"id": "ac3", "description": "New regression test added for the vulnerability", "type": "test", "weight": 0.3, "pass_threshold": 1.0},
			},
			"scoring_weights_template": map[string]interface{}{
				"correctness": 0.4, "security": 0.3, "performance": 0.15, "style": 0.1, "test_coverage": 0.05,
			},
			"known_issues_template": []map[string]interface{}{
				{"file": "", "line": 0, "issue": "Add known issues here", "severity": "high", "priority": "high"},
			},
			"example_repository": "https://github.com/example/vulnerable-app",
		},
		{
			"name":                       "performance-optimization",
			"description":                "Template for performance optimization challenges",
			"category":                   "performance",
			"default_severity":           "medium",
			"default_time_limit_seconds": 300,
			"default_max_turns":          8,
			"acceptance_criteria_template": []map[string]interface{}{
				{"id": "ac1", "description": "Execution time improved by at least 20%", "type": "metric", "weight": 0.5, "pass_threshold": 1.2},
				{"id": "ac2", "description": "Memory usage reduced or stable", "type": "metric", "weight": 0.3, "pass_threshold": 1.0},
				{"id": "ac3", "description": "All existing tests pass", "type": "test", "weight": 0.2, "pass_threshold": 1.0},
			},
			"scoring_weights_template": map[string]interface{}{
				"correctness": 0.3, "security": 0.1, "performance": 0.4, "style": 0.1, "test_coverage": 0.1,
			},
			"known_issues_template": []map[string]interface{}{
				{"file": "", "line": 0, "issue": "Identify performance bottlenecks", "severity": "medium", "priority": "high"},
			},
			"example_repository": "https://github.com/example/slow-app",
		},
		{
			"name":                       "refactor-legacy",
			"description":                "Template for legacy code refactoring",
			"category":                   "refactor",
			"default_severity":           "medium",
			"default_time_limit_seconds": 400,
			"default_max_turns":          12,
			"acceptance_criteria_template": []map[string]interface{}{
				{"id": "ac1", "description": "Cyclomatic complexity reduced by 30%", "type": "metric", "weight": 0.3, "pass_threshold": 0.7},
				{"id": "ac2", "description": "Code duplication reduced by 50%", "type": "metric", "weight": 0.3, "pass_threshold": 0.5},
				{"id": "ac3", "description": "All existing tests pass", "type": "test", "weight": 0.4, "pass_threshold": 1.0},
			},
			"scoring_weights_template": map[string]interface{}{
				"correctness": 0.35, "security": 0.1, "performance": 0.15, "style": 0.3, "test_coverage": 0.1,
			},
			"known_issues_template": []map[string]interface{}{
				{"file": "", "line": 0, "issue": "Identify code smells", "severity": "medium", "priority": "high"},
			},
			"example_repository": "https://github.com/example/legacy-app",
		},
		{
			"name":                       "add-test-coverage",
			"description":                "Template for adding test coverage",
			"category":                   "testing",
			"default_severity":           "low",
			"default_time_limit_seconds": 300,
			"default_max_turns":          10,
			"acceptance_criteria_template": []map[string]interface{}{
				{"id": "ac1", "description": "Line coverage increased to 80%+", "type": "metric", "weight": 0.4, "pass_threshold": 0.8},
				{"id": "ac2", "description": "Branch coverage increased to 70%+", "type": "metric", "weight": 0.3, "pass_threshold": 0.7},
				{"id": "ac3", "description": "New tests are meaningful (mutation score > 80%)", "type": "test", "weight": 0.3, "pass_threshold": 0.8},
			},
			"scoring_weights_template": map[string]interface{}{
				"correctness": 0.3, "security": 0.05, "performance": 0.05, "style": 0.1, "test_coverage": 0.5,
			},
			"known_issues_template": []map[string]interface{}{
				{"file": "", "line": 0, "issue": "Identify untested code paths", "severity": "low", "priority": "medium"},
			},
			"example_repository": "https://github.com/example/untested-app",
		},
	}

	for _, t := range templates {
		acceptanceCriteria, _ := json.Marshal(t["acceptance_criteria_template"])
		scoringWeights, _ := json.Marshal(t["scoring_weights_template"])
		knownIssues, _ := json.Marshal(t["known_issues_template"])

		_, err := pool.Exec(ctx, `
			INSERT INTO challenge_templates (name, description, category, default_severity, default_time_limit_seconds, default_max_turns, acceptance_criteria_template, scoring_weights_template, known_issues_template, example_repository)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (name) DO NOTHING
		`,
			t["name"], t["description"], t["category"], t["default_severity"],
			t["default_time_limit_seconds"], t["default_max_turns"],
			acceptanceCriteria, scoringWeights, knownIssues, t["example_repository"],
		)
		if err != nil {
			log.Printf("Failed to insert template %s: %v", t["name"], err)
		}
	}

	return nil
}

func seedSnippets(ctx context.Context, pool *pgxpool.Pool) error {
	snippets := []map[string]interface{}{
		{
			"title":       "Binary Search",
			"description": "Iterative binary search implementation",
			"language":    "python",
			"code":        "def binary_search(arr, target):\n    left, right = 0, len(arr) - 1\n    while left <= right:\n        mid = (left + right) // 2\n        if arr[mid] == target:\n            return mid\n        elif arr[mid] < target:\n            left = mid + 1\n        else:\n            right = mid - 1\n    return -1",
			"tags":        []string{"search", "algorithm", "divide-conquer"},
			"complexity":  "basic",
			"author":      "cp-algorithms",
			"is_official": true,
		},
		{
			"title":       "Quick Sort",
			"description": "In-place quick sort with random pivot",
			"language":    "python",
			"code":        "import random\n\ndef quick_sort(arr, low=0, high=None):\n    if high is None:\n        high = len(arr) - 1\n    if low < high:\n        pivot_idx = random.randint(low, high)\n        arr[pivot_idx], arr[high] = arr[high], arr[pivot_idx]\n        pi = partition(arr, low, high)\n        quick_sort(arr, low, pi - 1)\n        quick_sort(arr, pi + 1, high)\n\ndef partition(arr, low, high):\n    pivot = arr[high]\n    i = low - 1\n    for j in range(low, high):\n        if arr[j] <= pivot:\n            i += 1\n            arr[i], arr[j] = arr[j], arr[i]\n    arr[i + 1], arr[high] = arr[high], arr[i + 1]\n    return i + 1",
			"tags":        []string{"sort", "algorithm", "divide-conquer"},
			"complexity":  "intermediate",
			"author":      "cp-algorithms",
			"is_official": true,
		},
		{
			"title":       "Dijkstra's Algorithm",
			"description": "Shortest path in weighted graph",
			"language":    "python",
			"code":        "import heapq\n\ndef dijkstra(graph, start):\n    distances = {node: float('inf') for node in graph}\n    distances[start] = 0\n    pq = [(0, start)]\n\n    while pq:\n        dist, node = heapq.heappop(pq)\n        if dist > distances[node]:\n            continue\n        for neighbor, weight in graph[node].items():\n            new_dist = dist + weight\n            if new_dist < distances[neighbor]:\n                distances[neighbor] = new_dist\n                heapq.heappush(pq, (new_dist, neighbor))\n    return distances",
			"tags":        []string{"graph", "shortest-path", "dijkstra"},
			"complexity":  "advanced",
			"author":      "cp-algorithms",
			"is_official": true,
		},
		{
			"title":       "Binary Search Tree Insert",
			"description": "BST insertion in C++",
			"language":    "cpp",
			"code":        "struct TreeNode {\n    int val;\n    TreeNode *left, *right;\n    TreeNode(int x) : val(x), left(nullptr), right(nullptr) {}\n};\n\nTreeNode* insert(TreeNode* root, int val) {\n    if (!root) return new TreeNode(val);\n    if (val < root->val)\n        root->left = insert(root->left, val);\n    else\n        root->right = insert(root->right, val);\n    return root;\n}",
			"tags":        []string{"bst", "tree", "insert"},
			"complexity":  "basic",
			"author":      "awesome-algorithms",
			"is_official": true,
		},
		{
			"title":       "LRU Cache",
			"description": "LRU Cache implementation with O(1) operations",
			"language":    "java",
			"code":        "import java.util.*;\n\nclass LRUCache {\n    private final int capacity;\n    private final Map<Integer, Node> cache;\n    private final Node head, tail;\n\n    public LRUCache(int capacity) {\n        this.capacity = capacity;\n        this.cache = new HashMap<>();\n        this.head = new Node(0, 0);\n        this.tail = new Node(0, 0);\n        head.next = tail;\n        tail.prev = head;\n    }\n\n    public int get(int key) {\n        Node node = cache.get(key);\n        if (node == null) return -1;\n        moveToHead(node);\n        return node.value;\n    }\n\n    public void put(int key, int value) {\n        Node node = cache.get(key);\n        if (node == null) {\n            node = new Node(key, value);\n            cache.put(key, node);\n            addToHead(node);\n            if (cache.size() > capacity) {\n                Node removed = removeTail();\n                cache.remove(removed.key);\n            }\n        } else {\n            node.value = value;\n            moveToHead(node);\n        }\n    }\n\n    private void addToHead(Node node) {\n        node.next = head.next;\n        node.prev = head;\n        head.next.prev = node;\n        head.next = node;\n    }\n\n    private void removeNode(Node node) {\n        node.prev.next = node.next;\n        node.next.prev = node.prev;\n    }\n\n    private void moveToHead(Node node) {\n        removeNode(node);\n        addToHead(node);\n    }\n\n    private Node removeTail() {\n        Node node = tail.prev;\n        removeNode(node);\n        return node;\n    }\n\n    static class Node {\n        int key, value;\n        Node prev, next;\n        Node(int key, int value) { this.key = key; this.value = value; }\n    }\n}",
			"tags":        []string{"cache", "lru", "design-pattern"},
			"complexity":  "advanced",
			"author":      "leetcode-patterns",
			"is_official": true,
		},
		{
			"title":       "Debounce Function",
			"description": "Debounce utility for JavaScript/TypeScript",
			"language":    "typescript",
			"code":        "export function debounce<T extends (...args: unknown[]) => unknown>(\n  func: T,\n  wait: number\n): (...args: Parameters<T>) => void {\n  let timeoutId: ReturnType<typeof setTimeout> | null = null;\n\n  return (...args: Parameters<T>) => {\n    if (timeoutId) clearTimeout(timeoutId);\n    timeoutId = setTimeout(() => func(...args), wait);\n  };\n}\n\n// Usage\nconst debouncedSearch = debounce((query: string) => {\n  console.log('Searching for:', query);\n}, 300);",
			"tags":        []string{"utility", "debounce", "async"},
			"complexity":  "basic",
			"author":      "programming-idioms",
			"is_official": true,
		},
	}

	for _, s := range snippets {
		tags, _ := json.Marshal(s["tags"])
		_, err := pool.Exec(ctx, `
			INSERT INTO snippets (title, description, language, code, tags, complexity, author, is_official)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT DO NOTHING
		`,
			s["title"], s["description"], s["language"], s["code"],
			tags, s["complexity"], s["author"], s["is_official"],
		)
		if err != nil {
			log.Printf("Failed to insert snippet %s: %v", s["title"], err)
		}
	}

	return nil
}
