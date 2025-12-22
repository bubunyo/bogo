package bogo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// Test data structures for benchmarking
type SimpleUser struct {
	ID       int64     `json:"id" msgpack:"id"`
	Username string    `json:"username" msgpack:"username"`
	Email    string    `json:"email" msgpack:"email"`
	Active   bool      `json:"active" msgpack:"active"`
	Balance  float64   `json:"balance" msgpack:"balance"`
	Created  time.Time `json:"created" msgpack:"created"`
}

type ComplexData struct {
	Meta struct {
		Version   string    `json:"version" msgpack:"version"`
		Timestamp time.Time `json:"timestamp" msgpack:"timestamp"`
		RequestID string    `json:"request_id" msgpack:"request_id"`
	} `json:"meta" msgpack:"meta"`

	User struct {
		ID          int64          `json:"id" msgpack:"id"`
		Username    string         `json:"username" msgpack:"username"`
		Email       string         `json:"email" msgpack:"email"`
		IsVerified  bool           `json:"is_verified" msgpack:"is_verified"`
		IsPremium   bool           `json:"is_premium" msgpack:"is_premium"`
		Balance     float64        `json:"balance" msgpack:"balance"`
		ProfilePic  []byte         `json:"profile_pic" msgpack:"profile_pic"`
		Tags        []string       `json:"tags" msgpack:"tags"`
		LoginTimes  []int64        `json:"login_times" msgpack:"login_times"`
		Scores      []float64      `json:"scores" msgpack:"scores"`
		Preferences map[string]any `json:"preferences" msgpack:"preferences"`
	} `json:"user" msgpack:"user"`

	Analytics struct {
		PageViews    []int64            `json:"page_views" msgpack:"page_views"`
		ClickEvents  []string           `json:"click_events" msgpack:"click_events"`
		SessionTimes []float64          `json:"session_times" msgpack:"session_times"`
		Metrics      map[string]float64 `json:"metrics" msgpack:"metrics"`
		FeatureFlags map[string]bool    `json:"feature_flags" msgpack:"feature_flags"`
	} `json:"analytics" msgpack:"analytics"`
}

// Create test data instances
func createSimpleUser() SimpleUser {
	return SimpleUser{
		ID:       12345,
		Username: "john_doe_2024",
		Email:    "john.doe@example.com",
		Active:   true,
		Balance:  1234.56,
		Created:  time.Unix(1703030400, 0),
	}
}

func createComplexData() ComplexData {
	data := ComplexData{}

	// Meta section
	data.Meta.Version = "v2.1.3"
	data.Meta.Timestamp = time.Unix(1703030400, 0)
	data.Meta.RequestID = "req_abc123def456"

	// User section
	data.User.ID = 987654321
	data.User.Username = "alice_developer"
	data.User.Email = "alice@techcorp.com"
	data.User.IsVerified = true
	data.User.IsPremium = false
	data.User.Balance = 2847.92
	data.User.ProfilePic = []byte("fake_compressed_image_data_here_12345")
	data.User.Tags = []string{"developer", "premium_trial", "early_adopter", "beta_tester"}
	data.User.LoginTimes = []int64{1703001600, 1702915200, 1702828800, 1702742400, 1702656000}
	data.User.Scores = []float64{85.5, 90.2, 78.8, 92.1, 88.7, 95.3, 82.4}
	data.User.Preferences = map[string]any{
		"theme":           "dark",
		"notifications":   true,
		"language":        "en-US",
		"timezone_offset": -5,
		"beta_features":   true,
	}

	// Analytics section
	data.Analytics.PageViews = []int64{1520, 1687, 1734, 1823, 1756, 1892, 2103}
	data.Analytics.ClickEvents = []string{
		"button_click", "link_click", "menu_open", "search_query",
		"file_download", "share_action", "profile_edit", "settings_change",
	}
	data.Analytics.SessionTimes = []float64{45.2, 67.8, 34.1, 89.5, 23.7, 56.9, 78.3}
	data.Analytics.Metrics = map[string]float64{
		"engagement_rate": 78.5,
		"conversion_rate": 12.3,
		"retention_rate":  89.7,
		"satisfaction":    4.2,
		"nps_score":       67.0,
	}
	data.Analytics.FeatureFlags = map[string]bool{
		"new_ui":          true,
		"beta_analytics":  false,
		"ai_assistant":    true,
		"advanced_export": false,
		"real_time_sync":  true,
	}

	return data
}

// Convert ComplexData to map[string]any for Bogo (since it doesn't support struct tags)
func complexDataToMap(data ComplexData) map[string]any {
	return map[string]any{
		"meta": map[string]any{
			"version":    data.Meta.Version,
			"timestamp":  data.Meta.Timestamp,
			"request_id": data.Meta.RequestID,
		},
		"user": map[string]any{
			"id":          int64(data.User.ID),
			"username":    data.User.Username,
			"email":       data.User.Email,
			"is_verified": data.User.IsVerified,
			"is_premium":  data.User.IsPremium,
			"balance":     data.User.Balance,
			"profile_pic": data.User.ProfilePic,
			"tags":        data.User.Tags,
			"login_times": data.User.LoginTimes,
			"scores":      data.User.Scores,
			"preferences": data.User.Preferences,
		},
		"analytics": map[string]any{
			"page_views":    data.Analytics.PageViews,
			"click_events":  data.Analytics.ClickEvents,
			"session_times": data.Analytics.SessionTimes,
			"metrics":       data.Analytics.Metrics,
			"feature_flags": data.Analytics.FeatureFlags,
		},
	}
}

func simpleUserToMap(user SimpleUser) map[string]any {
	return map[string]any{
		"id":       int64(user.ID),
		"username": user.Username,
		"email":    user.Email,
		"active":   user.Active,
		"balance":  user.Balance,
		"created":  user.Created,
	}
}

// === SIMPLE DATA BENCHMARKS ===

func BenchmarkSimpleData(b *testing.B) {
	user := createSimpleUser()
	userMap := simpleUserToMap(user)

	b.Run("Serialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := json.Marshal(user)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Marshal(userMap)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := msgpack.Marshal(user)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	// Prepare data for deserialize benchmarks
	jsonData, _ := json.Marshal(user)
	bogoData, _ := Marshal(userMap)
	msgpackData, _ := msgpack.Marshal(user)

	b.Logf("Simple data sizes - JSON: %d bytes, Bogo: %d bytes, MessagePack: %d bytes",
		len(jsonData), len(bogoData), len(msgpackData))

	b.Run("Deserialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result SimpleUser
				err := json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := Unmarshal(bogoData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result SimpleUser
				err := msgpack.Unmarshal(msgpackData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

// === COMPLEX DATA BENCHMARKS ===

func BenchmarkComplexData(b *testing.B) {
	complexData := createComplexData()
	complexMap := complexDataToMap(complexData)

	b.Run("Serialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := json.Marshal(complexData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Marshal(complexMap)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := msgpack.Marshal(complexData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	// Prepare data for deserialize benchmarks
	jsonData, _ := json.Marshal(complexData)
	bogoData, _ := Marshal(complexMap)
	msgpackData, _ := msgpack.Marshal(complexData)

	b.Logf("Complex data sizes - JSON: %d bytes, Bogo: %d bytes, MessagePack: %d bytes",
		len(jsonData), len(bogoData), len(msgpackData))

	b.Run("Deserialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result ComplexData
				err := json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := Unmarshal(bogoData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result ComplexData
				err := msgpack.Unmarshal(msgpackData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

// === ARRAY DATA BENCHMARKS ===

func BenchmarkArrayData(b *testing.B) {
	// Test with different array types
	stringArray := make([]string, 100)
	intArray := make([]int64, 100)
	floatArray := make([]float64, 100)

	// Fill arrays with test data
	for i := 0; i < 100; i++ {
		stringArray[i] = "test_string_" + string(rune(i+65)) // A, B, C, etc.
		intArray[i] = int64(i * 100)
		floatArray[i] = float64(i) * 3.14159
	}

	testData := map[string]any{
		"strings":  stringArray,
		"integers": intArray,
		"floats":   floatArray,
		"metadata": map[string]any{
			"count": len(stringArray),
			"type":  "test_arrays",
		},
	}

	b.Run("Arrays_Serialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := json.Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := msgpack.Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	// Prepare data for deserialize benchmarks
	jsonData, _ := json.Marshal(testData)
	bogoData, _ := Marshal(testData)
	msgpackData, _ := msgpack.Marshal(testData)

	b.Logf("Array data sizes - JSON: %d bytes, Bogo: %d bytes, MessagePack: %d bytes",
		len(jsonData), len(bogoData), len(msgpackData))

	b.Run("Arrays_Deserialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := Unmarshal(bogoData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := msgpack.Unmarshal(msgpackData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

// === BINARY DATA BENCHMARKS ===

func BenchmarkBinaryData(b *testing.B) {
	// Create test data with binary blobs of different sizes
	smallBlob := make([]byte, 100)
	mediumBlob := make([]byte, 1000)
	largeBlob := make([]byte, 10000)

	// Fill with test patterns
	for i := range smallBlob {
		smallBlob[i] = byte(i % 256)
	}
	for i := range mediumBlob {
		mediumBlob[i] = byte(i % 256)
	}
	for i := range largeBlob {
		largeBlob[i] = byte(i % 256)
	}

	testData := map[string]any{
		"small_blob":  smallBlob,
		"medium_blob": mediumBlob,
		"large_blob":  largeBlob,
		"metadata": map[string]any{
			"small_size":  len(smallBlob),
			"medium_size": len(mediumBlob),
			"large_size":  len(largeBlob),
		},
	}

	b.Run("Binary_Serialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := json.Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := msgpack.Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	// Prepare data for deserialize benchmarks
	jsonData, _ := json.Marshal(testData)
	bogoData, _ := Marshal(testData)
	msgpackData, _ := msgpack.Marshal(testData)

	b.Logf("Binary data sizes - JSON: %d bytes, Bogo: %d bytes, MessagePack: %d bytes",
		len(jsonData), len(bogoData), len(msgpackData))

	b.Run("Binary_Deserialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := Unmarshal(bogoData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := msgpack.Unmarshal(msgpackData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

// === COMPREHENSIVE BENCHMARK ===

func BenchmarkComprehensive(b *testing.B) {
	// Use the realistic data from the comprehensive test
	testData := map[string]any{
		"api_version":   "v2.1.3",
		"success":       true,
		"timestamp":     time.Unix(1703030400, 0),
		"response_time": 0.125,
		"user": map[string]any{
			"id":          int64(987654321),
			"username":    "alice_developer",
			"email":       "alice@techcorp.com",
			"is_verified": true,
			"is_premium":  false,
			"balance":     2847.92,
			"avatar_data": []byte("compressed_image_data_here_12345"),
			"login_times": []int64{1703001600, 1702915200, 1702828800},
			"scores":      []float64{85.5, 90.2, 78.8, 92.1, 88.7},
			"tags":        []string{"developer", "premium_trial", "early_adopter"},
			"preferences": map[string]any{
				"theme":           "dark",
				"notifications":   true,
				"language":        "en-US",
				"timezone_offset": -5,
				"beta_features":   true,
			},
		},
		"analytics": map[string]any{
			"page_views":    []int64{1520, 1687, 1734, 1823, 1756},
			"session_times": []float64{45.2, 67.8, 34.1, 89.5, 23.7},
			"metrics": map[string]any{
				"engagement_rate": 78.5,
				"conversion_rate": 12.3,
				"nps_score":       67.0,
			},
			"feature_flags": map[string]any{
				"new_ui":         true,
				"beta_analytics": false,
				"ai_assistant":   true,
			},
		},
	}

	b.Run("Comprehensive_Serialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := json.Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := msgpack.Marshal(testData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	// Prepare data for deserialize benchmarks
	jsonData, _ := json.Marshal(testData)
	bogoData, _ := Marshal(testData)
	msgpackData, _ := msgpack.Marshal(testData)

	b.Logf("Comprehensive data sizes - JSON: %d bytes, Bogo: %d bytes, MessagePack: %d bytes",
		len(jsonData), len(bogoData), len(msgpackData))

	b.Run("Comprehensive_Deserialize", func(b *testing.B) {
		b.Run("JSON", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Bogo", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := Unmarshal(bogoData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("MessagePack", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result map[string]any
				err := msgpack.Unmarshal(msgpackData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

