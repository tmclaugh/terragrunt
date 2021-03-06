package config

import (
	"testing"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/test/helpers"
	"fmt"
)

func TestPathRelativeToInclude(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		include           *IncludeConfig
		terragruntOptions options.TerragruntOptions
		expectedPath      string
	}{
		{
			nil,
			options.TerragruntOptions{TerragruntConfigPath: helpers.RootFolder + "child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			".",
		},
		{
			&IncludeConfig{Path: "../" + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: helpers.RootFolder + "child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child",
		},
		{
			&IncludeConfig{Path: helpers.RootFolder + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: helpers.RootFolder + "child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child",
		},
		{
			&IncludeConfig{Path: "../../../" + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: helpers.RootFolder + "child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child/sub-child/sub-sub-child",
		},
		{
			&IncludeConfig{Path: helpers.RootFolder + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: helpers.RootFolder + "child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child/sub-child/sub-sub-child",
		},
		{
			&IncludeConfig{Path: "../../other-child/" + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: helpers.RootFolder + "child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"../child/sub-child",
		},
		{
			&IncludeConfig{Path: "../../" + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: "../child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child/sub-child",
		},
		{
			&IncludeConfig{Path: "${find_in_parent_folders()}"},
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child/sub-child",
		},
	}

	for _, testCase := range testCases {
		actualPath, actualErr := pathRelativeToInclude(testCase.include, &testCase.terragruntOptions)
		assert.Nil(t, actualErr, "For include %v and options %v, unexpected error: %v", testCase.include, testCase.terragruntOptions, actualErr)
		assert.Equal(t, testCase.expectedPath, actualPath, "For include %v and options %v", testCase.include, testCase.terragruntOptions)
	}
}

func TestFindInParentFolders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		terragruntOptions options.TerragruntOptions
		expectedPath      string
		expectedErr       error
	}{
		{
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/terragrunt-in-root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"../../../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/no-terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			ParentTerragruntConfigNotFound("../test/fixture-parent-folders/no-terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath),
		},
		{
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/multiple-terragrunt-in-parents/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/multiple-terragrunt-in-parents/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/multiple-terragrunt-in-parents/child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.TerragruntOptions{TerragruntConfigPath: "/", NonInteractive: true},
			"",
			ParentTerragruntConfigNotFound("/"),
		},
		{
			options.TerragruntOptions{TerragruntConfigPath: "/fake/path", NonInteractive: true},
			"",
			ParentTerragruntConfigNotFound("/fake/path"),
		},
	}

	for _, testCase := range testCases {
		actualPath, actualErr := findInParentFolders(&testCase.terragruntOptions)
		if testCase.expectedErr != nil {
			assert.True(t, errors.IsError(actualErr, testCase.expectedErr), "For options %v, expected error %v but got error %v", testCase.terragruntOptions, testCase.expectedErr, actualErr)
		} else {
			assert.Nil(t, actualErr, "For options %v, unexpected error: %v", testCase.terragruntOptions, actualErr)
			assert.Equal(t, testCase.expectedPath, actualPath, "For options %v", testCase.terragruntOptions)
		}
	}
}

func TestResolveTerragruntInterpolation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		str               string
		include           *IncludeConfig
		terragruntOptions options.TerragruntOptions
		expectedOut       string
		expectedErr       error
	}{
		{
			"${path_relative_to_include()}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			".",
			nil,
		},
		{
			"${path_relative_to_include()}",
			&IncludeConfig{Path: "../" + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child",
			nil,
		},
		{
			"${path_relative_to_include()}",
			&IncludeConfig{Path: "${find_in_parent_folders()}"},
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child/sub-child",
			nil,
		},
		{
			"${find_in_parent_folders()}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"../../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			"${find_in_parent_folders()}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/no-terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			ParentTerragruntConfigNotFound("../test/fixture-parent-folders/no-terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath),
		},
		{
			"${find_in_parent_folders}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidInterpolationSyntax("${find_in_parent_folders}"),
		},
		{
			"{find_in_parent_folders()}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidInterpolationSyntax("{find_in_parent_folders()}"),
		},
		{
			"find_in_parent_folders()",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidInterpolationSyntax("find_in_parent_folders()"),
		},
	}

	for _, testCase := range testCases {
		actualOut, actualErr := resolveTerragruntInterpolation(testCase.str, testCase.include, &testCase.terragruntOptions)
		if testCase.expectedErr != nil {
			assert.True(t, errors.IsError(actualErr, testCase.expectedErr), "For string '%s' include %v and options %v, expected error %v but got error %v", testCase.str, testCase.include, testCase.terragruntOptions, testCase.expectedErr, actualErr)
		} else {
			assert.Nil(t, actualErr, "For string '%s' include %v and options %v, unexpected error: %v", testCase.str, testCase.include, testCase.terragruntOptions, actualErr)
			assert.Equal(t, testCase.expectedOut, actualOut, "For string '%s' include %v and options %v", testCase.str, testCase.include, testCase.terragruntOptions)
		}
	}
}

func TestResolveTerragruntConfigString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		str               string
		include           *IncludeConfig
		terragruntOptions options.TerragruntOptions
		expectedOut       string
		expectedErr       error
	}{
		{
			"",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			nil,
		},
		{
			"foo bar",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"foo bar",
			nil,
		},
		{
			"$foo {bar}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"$foo {bar}",
			nil,
		},
		{
			"${path_relative_to_include()}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			".",
			nil,
		},
		{
			"${path_relative_to_include()}",
			&IncludeConfig{Path: "../" + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child",
			nil,
		},
		{
			"${path_relative_to_include()}",
			&IncludeConfig{Path: "${find_in_parent_folders()}"},
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"child/sub-child",
			nil,
		},
		{
			"foo/${path_relative_to_include()}/bar",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"foo/./bar",
			nil,
		},
		{
			"foo/${path_relative_to_include()}/bar",
			&IncludeConfig{Path: "../" + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"foo/child/bar",
			nil,
		},
		{
			"foo/${path_relative_to_include()}/bar/${path_relative_to_include()}",
			&IncludeConfig{Path: "../" + DefaultTerragruntConfigPath},
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"foo/child/bar/child",
			nil,
		},
		{
			"${find_in_parent_folders()}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"../../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			"foo/${find_in_parent_folders()}/bar",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			fmt.Sprintf("foo/../../%s/bar", DefaultTerragruntConfigPath),
			nil,
		},
		{
			"${find_in_parent_folders()}",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "../test/fixture-parent-folders/no-terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			ParentTerragruntConfigNotFound("../test/fixture-parent-folders/no-terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath),
		},
		{
			"foo/${unknown}/bar",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidInterpolationSyntax("${unknown}"),
		},
	}

	for _, testCase := range testCases {
		actualOut, actualErr := ResolveTerragruntConfigString(testCase.str, testCase.include, &testCase.terragruntOptions)
		if testCase.expectedErr != nil {
			assert.True(t, errors.IsError(actualErr, testCase.expectedErr), "For string '%s' include %v and options %v, expected error %v but got error %v", testCase.str, testCase.include, testCase.terragruntOptions, testCase.expectedErr, actualErr)
		} else {
			assert.Nil(t, actualErr, "For string '%s' include %v and options %v, unexpected error: %v", testCase.str, testCase.include, testCase.terragruntOptions, actualErr)
			assert.Equal(t, testCase.expectedOut, actualOut, "For string '%s' include %v and options %v", testCase.str, testCase.include, testCase.terragruntOptions)
		}
	}
}

func TestResolveEnvInterpolationConfigString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		str            string
		include        *IncludeConfig
		terragruntOptions options.TerragruntOptions
		expectedOut    string
		expectedErr    error
	}{
		{
			"foo/${get_env()}/bar",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidFunctionParameters(""),
		},
		{
			"foo/${get_env(Invalid Parameters)}/bar",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidFunctionParameters("Invalid Parameters"),
		},
		{
			"foo/${get_env('env','')}/bar",
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidFunctionParameters("'env',''"),
		},
		{
			`foo/${get_env("","")}/bar`,
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidFunctionParameters(`"",""`),
		},
		{
			`foo/${get_env(   ""    ,   ""    )}/bar`,
			nil,
			options.TerragruntOptions{TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath, NonInteractive: true},
			"",
			InvalidFunctionParameters(`   ""    ,   ""    `),
		},
		{
			`foo/${get_env("TEST_ENV_TERRAGRUNT_HIT","")}/bar`,
			nil,
			options.TerragruntOptions{
				TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath,
				NonInteractive: true,
				Env: map[string]string{"TEST_ENV_TERRAGRUNT_OTHER": "SOMETHING"},
			},
			"foo//bar",
			nil,
		},
		{
			`foo/${get_env(    "TEST_ENV_TERRAGRUNT_HIT"   ,   ""   )}/bar`,
			nil,
			options.TerragruntOptions{
				TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath,
				NonInteractive: true,
				Env: map[string]string{"TEST_ENV_TERRAGRUNT_OTHER": "SOMETHING"},
			},
			"foo//bar",
			nil,
		},
		{
		`foo/${get_env("TEST_ENV_
TERRAGRUNT_HIT","")}/bar`,
		nil,
		options.TerragruntOptions{
			TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath,
			NonInteractive: true,
			Env: map[string]string{"TEST_ENV_TERRAGRUNT_OTHER": "SOMETHING"},
		},
			`foo/${get_env("TEST_ENV_
TERRAGRUNT_HIT","")}/bar`,
		nil,
		},
		{
			`foo/${get_env("TEST_ENV_TERRAGRUNT_HIT","DEFAULT")}/bar`,
			nil,
			options.TerragruntOptions{
				TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath,
				NonInteractive: true,
				Env: map[string]string{"TEST_ENV_TERRAGRUNT_OTHER": "SOMETHING"},
			},
			"foo/DEFAULT/bar",
			nil,
		},
		{
			`foo/${get_env(    "TEST_ENV_TERRAGRUNT_HIT      "   ,   "default"   )}/bar`,
			nil,
			options.TerragruntOptions{
				TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath,
				NonInteractive: true,
				Env: map[string]string{"TEST_ENV_TERRAGRUNT_HIT": "environment hit  "},
			},
			"foo/environment hit  /bar",
			nil,
		},
		{
			`foo/${get_env("TEST_ENV_TERRAGRUNT_HIT","default")}/bar`,
			nil,
			options.TerragruntOptions{
				TerragruntConfigPath: "/root/child/" + DefaultTerragruntConfigPath,
				NonInteractive: true,
				Env: map[string]string{"TEST_ENV_TERRAGRUNT_HIT": "HIT"},
			},
			"foo/HIT/bar",
			nil,
		},
	}

	for _, testCase := range testCases {
		actualOut, actualErr := ResolveTerragruntConfigString(testCase.str, testCase.include, &testCase.terragruntOptions)
		if testCase.expectedErr != nil {
			assert.True(t, errors.IsError(actualErr, testCase.expectedErr), "For string '%s' include %v and options %v, expected error %v but got error %v and output %v", testCase.str, testCase.include, testCase.terragruntOptions, testCase.expectedErr, actualErr, actualOut)
		} else {
			assert.Nil(t, actualErr, "For string '%s' include %v and options %v, unexpected error: %v", testCase.str, testCase.include, testCase.terragruntOptions, actualErr)
			assert.Equal(t, testCase.expectedOut, actualOut, "For string '%s' include %v and options %v", testCase.str, testCase.include, testCase.terragruntOptions)
		}
	}
}