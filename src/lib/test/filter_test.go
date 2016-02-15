package sybil_test

import sybil "../"

import "testing"
import "fmt"
import "math/rand"
import "strconv"
import "math"
import "strings"

func TestFilters(test *testing.T) {
	delete_test_db()
	sybil.CHUNK_SIZE = 100

	if testing.Short() {
		test.Skip("Skipping test in short mode")
		return
	}

	sybil.TEST_MODE = true

	BLOCK_COUNT := 3
	COUNT := sybil.CHUNK_SIZE * BLOCK_COUNT
	t := sybil.GetTable(TEST_TABLE_NAME)

	total_age := int64(0)
	for i := 0; i < COUNT; i++ {
		r := t.NewRecord()
		r.AddIntField("id", int64(i))
		age := int64(rand.Intn(20)) + 10
		total_age += age
		r.AddIntField("age", age)
		r.AddStrField("age_str", strconv.FormatInt(int64(age), 10))
	}

	avg_age := float64(total_age) / float64(COUNT)

	t.SaveRecords()

	nt := sybil.GetTable(TEST_TABLE_NAME)
	loadSpec := sybil.NewLoadSpec()
	loadSpec.LoadAllColumns = true
	loadSpec.Str("age_str")
	loadSpec.Int("id")
	loadSpec.Int("age")

	filters := []sybil.Filter{}
	filters = append(filters, nt.IntFilter("age", "eq", 20))

	aggs := []sybil.Aggregation{}
	groupings := []sybil.Grouping{}
	aggs = append(aggs, nt.Aggregation("age", "avg"))

	querySpec := sybil.QuerySpec{Groups: groupings, Filters: filters, Aggregations: aggs}
	querySpec.Punctuate()

	nt.LoadRecords(&loadSpec)
	nt.MatchAndAggregate(&querySpec)

	// Test Filtering to 20..
	if len(querySpec.Results) <= 0 {
		test.Error("Int Filter for age 20 returned no results")
	}

	for k, v := range querySpec.Results {
		k = strings.Replace(k, sybil.GROUP_DELIMITER, "", 1)

		if math.Abs(20-float64(v.Hists["age"].Avg)) > 0.1 {
			fmt.Println("ACC", v.Hists["age"].Avg)
			test.Error("GROUP BY YIELDED UNEXPECTED RESULTS", k, avg_age, v.Hists["age"].Avg)
		}
	}
	fmt.Println("RESULTS", len(querySpec.Results))

	//
	filters = []sybil.Filter{}
	filters = append(filters, nt.StrFilter("age_str", "re", "20"))
	nt.MatchAndAggregate(&querySpec)

	if len(querySpec.Results) <= 0 {
		test.Error("Str Filter for age 20 returned no results")
	}

	for k, v := range querySpec.Results {
		k = strings.Replace(k, sybil.GROUP_DELIMITER, "", 1)

		if math.Abs(20-float64(v.Hists["age"].Avg)) > 0.1 {
			fmt.Println("ACC", v.Hists["age"].Avg)
			test.Error("GROUP BY YIELDED UNEXPECTED RESULTS", k, avg_age, v.Hists["age"].Avg)
		}
	}
	fmt.Println("RESULTS", len(querySpec.Results))
	delete_test_db()

}
