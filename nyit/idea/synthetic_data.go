package main

import (
    "encoding/json"
    "math/rand"
    "time"
)

type SyntheticTrafficGenerator struct {
    attackPatterns map[string]AttackPattern
}

type AttackPattern struct {
    Name            string
    PacketSizeRange [2]int
    DurationRange   [2]time.Duration
    PayloadPatterns []string
    Frequency       float64
}

func NewSyntheticTrafficGenerator() *SyntheticTrafficGenerator {
    return &SyntheticTrafficGenerator{
        attackPatterns: map[string]AttackPattern{
            "sql_injection": {
                Name:            "SQL Injection",
                PacketSizeRange: [2]int{200, 2000},
                DurationRange:   [2]time.Duration{time.Millisecond * 100, time.Second * 2},
                PayloadPatterns: []string{
                    "' OR 1=1--",
                    "UNION SELECT * FROM users--",
                    "; DROP TABLE users;--",
                },
                Frequency: 0.1,
            },
            "xss_attack": {
                Name:            "XSS Attack",
                PacketSizeRange: [2]int{150, 1500},
                DurationRange:   [2]time.Duration{time.Millisecond * 50, time.Second},
                PayloadPatterns: []string{
                    "<script>alert('XSS')</script>",
                    "javascript:alert(document.cookie)",
                    "<img src=x onerror=alert('XSS')>",
                },
                Frequency: 0.15,
            },
        },
    }
}

func (stg *SyntheticTrafficGenerator) GenerateTrafficData(numSamples int) []TrafficSample {
    var samples []TrafficSample
    
    for i := 0; i < numSamples; i++ {
        if rand.Float64() < 0.2 { // 20% attack traffic
            samples = append(samples, stg.generateAttackTraffic())
        } else {
            samples = append(samples, stg.generateNormalTraffic())
        }
    }
    
    return samples
}

type TrafficSample struct {
    Timestamp   time.Time `json:"timestamp"`
    SourceIP    string    `json:"source_ip"`
    DestIP      string    `json:"dest_ip"`
    Protocol    string    `json:"protocol"`
    PacketSize  int       `json:"packet_size"`
    Duration    int64     `json:"duration_ms"`
    Payload     string    `json:"payload"`
    IsAttack    bool      `json:"is_attack"`
    AttackType  string    `json:"attack_type,omitempty"`
}

func (stg *SyntheticTrafficGenerator) generateNormalTraffic() TrafficSample {
    return TrafficSample{
        Timestamp:  time.Now(),
        SourceIP:   generateRandomIP(),
        DestIP:     generateRandomIP(),
        Protocol:   randomChoice([]string{"TCP", "UDP", "HTTP", "HTTPS"}),
        PacketSize: rand.Intn(1400) + 100,
        Duration:   int64(rand.Intn(5000) + 100),
        Payload:    generateNormalPayload(),
        IsAttack:   false,
    }
}

func (stg *SyntheticTrafficGenerator) generateAttackTraffic() TrafficSample {
    attackTypes := make([]string, 0, len(stg.attackPatterns))
    for attackType := range stg.attackPatterns {
        attackTypes = append(attackTypes, attackType)
    }
    
    selectedAttack := randomChoice(attackTypes)
    pattern := stg.attackPatterns[selectedAttack]
    
    return TrafficSample{
        Timestamp:  time.Now(),
        SourceIP:   generateRandomIP(),
        DestIP:     generateRandomIP(),
        Protocol:   "HTTP",
        PacketSize: rand.Intn(pattern.PacketSizeRange[1]-pattern.PacketSizeRange[0]) + pattern.PacketSizeRange[0],
        Duration:   int64(pattern.DurationRange[0] + time.Duration(rand.Int63n(int64(pattern.DurationRange[1]-pattern.DurationRange[0])))),
        Payload:    randomChoice(pattern.PayloadPatterns),
        IsAttack:   true,
        AttackType: selectedAttack,
    }
} 