//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
}

func TestGetDomainStat_EmptyInput(t *testing.T) {
	result, err := GetDomainStat(bytes.NewBufferString(""), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{}, result)
}

func TestGetDomainStat_InvalidJSON(t *testing.T) {
	data := "not a json"
	_, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.Error(t, err)
}

func TestGetDomainStat_CaseInsensitiveDomain(t *testing.T) {
	data := `{"Id":1,"Name":"Test","Username":"test","Email":"user@BROWSEDRIVE.gov","Phone":"1","Password":"p","Address":"a"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
}

func TestGetDomainStat_SameDomainMultipleTimes(t *testing.T) {
	data := `{"Id":1,"Name":"A","Username":"a","Email":"a@test.com","Phone":"1","Password":"p","Address":"x"}
{"Id":2,"Name":"B","Username":"b","Email":"b@test.com","Phone":"2","Password":"p","Address":"x"}
{"Id":3,"Name":"C","Username":"c","Email":"c@test.com","Phone":"3","Password":"p","Address":"x"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"test.com": 3}, result)
}

func TestGetDomainStat_DomainNotSubstring(t *testing.T) {
	data := `{"Id":1,"Name":"A","Username":"a","Email":"a@test.com.ru","Phone":"1","Password":"p","Address":"x"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{}, result)
}

func TestGetDomainStat_MultipleDifferentDomains(t *testing.T) {
	data := `{"Id":1,"Name":"A","Username":"a","Email":"a@one.com","Phone":"1","Password":"p","Address":"x"}
{"Id":2,"Name":"B","Username":"b","Email":"b@two.com","Phone":"2","Password":"p","Address":"x"}
{"Id":3,"Name":"C","Username":"c","Email":"c@one.com","Phone":"3","Password":"p","Address":"x"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"one.com": 2, "two.com": 1}, result)
}
