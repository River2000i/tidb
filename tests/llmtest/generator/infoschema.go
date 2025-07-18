// Copyright 2025 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generator

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/pingcap/tidb/tests/llmtest/logger"
	"github.com/pingcap/tidb/tests/llmtest/testcase"
	"go.uber.org/zap"
)

type infoschemaGenerator struct {
}

// Name implements PromptGenerator.Name
func (g *infoschemaGenerator) Name() string {
	return "infoschema"
}

// Groups implements PromptGenerator.Groups
func (g *infoschemaGenerator) Groups() []string {
	return []string{
		// scalar functions
		"and", "cast", "or", ">=", "<=", "=", "!=", "<", ">", "|", "%", "not", "in", "like", "case"}
}

// GeneratePrompt implements PromptGenerator.GeneratePrompt
func (g *infoschemaGenerator) GeneratePrompt(group string, count int, existCases []*testcase.Case) []openai.ChatCompletionMessageParamUnion {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, 2)

	systemPrompt := `
	You are a professional QA engineer testing a new SQL database compatible with MySQL.
	You are tasked with testing the compatibility of the database with MySQL for a specific select operation for table information_schema.tables and information_schema.schemate.
	You should write the queries to cover the corner cases of the operation.
	The common cases are not needed.
	You should try to use this operation with different valid argument types to test the implicit type conversion.
	You should try to use this operation with NULL to test the behavior of NULL.
	Please return a valid JSON object with the key "queries" and an array of strings as the value. Be careful with the escape characters.
	You should avoid using NOW(), RAND() or any other functions that return different results on each call.
	You should pack the related DDL in the same query.
	You should CREATE and DROP the table before and after using it.
	The SELECT statement should have stable order.
	You should consider the table struct different of MySQL and TiDB.
	You should make sure you query's result can match multi tables/schema, such as 'where table_name like '%t%' can match two tables'.
	You should create multi table/schema in one query.
	here is mysql table struct:
	mysql> desc SCHEMATA;
	+----------------------------+------------------+------+-----+---------+-------+
	| Field                      | Type             | Null | Key | Default | Extra |
	+----------------------------+------------------+------+-----+---------+-------+
	| CATALOG_NAME               | varchar (64)      | YES  |     | NULL    |       |
	| SCHEMA_NAME                | varchar (64)      | YES  |     | NULL    |       |
	| DEFAULT_CHARACTER_SET_NAME | varchar (64)      | NO   |     | NULL    |       |
	| DEFAULT_COLLATION_NAME     | varchar (64)      | NO   |     | NULL    |       |
	| SQL_PATH                   | binary (0)        | YES  |     | NULL    |       |
	| DEFAULT_ENCRYPTION         | enum ('NO','YES') | NO   |     | NULL    |       |
	+----------------------------+------------------+------+-----+---------+-------+
	6 rows in set (0.00 sec)

	mysql> desc TABLES;
	+-----------------+--------------------------------------------------------------------+------+-----+---------+-------+
	| Field           | Type                                                               | Null | Key | Default | Extra |
	+-----------------+--------------------------------------------------------------------+------+-----+---------+-------+
	| TABLE_CATALOG   | varchar (64)                                                        | YES  |     | NULL    |       |
	| TABLE_SCHEMA    | varchar (64)                                                        | YES  |     | NULL    |       |
	| TABLE_NAME      | varchar (64)                                                        | YES  |     | NULL    |       |
	| TABLE_TYPE      | enum ('BASE TABLE','VIEW','SYSTEM VIEW')                            | NO   |     | NULL    |       |
	| ENGINE          | varchar (64)                                                        | YES  |     | NULL    |       |
	| VERSION         | int                                                                | YES  |     | NULL    |       |
	| ROW_FORMAT      | enum ('Fixed','Dynamic','Compressed','Redundant','Compact','Paged') | YES  |     | NULL    |       |
	| TABLE_ROWS      | bigint unsigned                                                    | YES  |     | NULL    |       |
	| AVG_ROW_LENGTH  | bigint unsigned                                                    | YES  |     | NULL    |       |
	| DATA_LENGTH     | bigint unsigned                                                    | YES  |     | NULL    |       |
	| MAX_DATA_LENGTH | bigint unsigned                                                    | YES  |     | NULL    |       |
	| INDEX_LENGTH    | bigint unsigned                                                    | YES  |     | NULL    |       |
	| DATA_FREE       | bigint unsigned                                                    | YES  |     | NULL    |       |
	| AUTO_INCREMENT  | bigint unsigned                                                    | YES  |     | NULL    |       |
	| CREATE_TIME     | timestamp                                                          | NO   |     | NULL    |       |
	| UPDATE_TIME     | datetime                                                           | YES  |     | NULL    |       |
	| CHECK_TIME      | datetime                                                           | YES  |     | NULL    |       |
	| TABLE_COLLATION | varchar (64)                                                        | YES  |     | NULL    |       |
	| CHECKSUM        | bigint                                                             | YES  |     | NULL    |       |
	| CREATE_OPTIONS  | varchar (256)                                                       | YES  |     | NULL    |       |
	| TABLE_COMMENT   | text                                                               | YES  |     | NULL    |       |
	+-----------------+--------------------------------------------------------------------+------+-----+---------+-------+
	21 rows in set (0.01 sec)

	here is tidb table struct
	mysql> desc schemata;
	+----------------------------+--------------+------+------+---------+-------+
	| Field                      | Type         | Null | Key  | Default | Extra |
	+----------------------------+--------------+------+------+---------+-------+
	| CATALOG_NAME               | varchar (512) | YES  |      | NULL    |       |
	| SCHEMA_NAME                | varchar (64)  | YES  |      | NULL    |       |
	| DEFAULT_CHARACTER_SET_NAME | varchar (64)  | YES  |      | NULL    |       |
	| DEFAULT_COLLATION_NAME     | varchar (32)  | YES  |      | NULL    |       |
	| SQL_PATH                   | varchar (512) | YES  |      | NULL    |       |
	| TIDB_PLACEMENT_POLICY_NAME | varchar (64)  | YES  |      | NULL    |       |
	+----------------------------+--------------+------+------+---------+-------+
	6 rows in set (0.00 sec)

	mysql> desc tables;
	+----------------------------+---------------+------+------+-------------+-------+
	| Field                      | Type          | Null | Key  | Default     | Extra |
	+----------------------------+---------------+------+------+-------------+-------+
	| TABLE_CATALOG              | varchar (512)  | YES  |      | NULL        |       |
	| TABLE_SCHEMA               | varchar (64)   | YES  |      | NULL        |       |
	| TABLE_NAME                 | varchar (64)   | YES  |      | NULL        |       |
	| TABLE_TYPE                 | varchar (64)   | YES  |      | NULL        |       |
	| ENGINE                     | varchar (64)   | YES  |      | NULL        |       |
	| VERSION                    | bigint        | YES  |      | NULL        |       |
	| ROW_FORMAT                 | varchar (10)   | YES  |      | NULL        |       |
	| TABLE_ROWS                 | bigint        | YES  |      | NULL        |       |
	| AVG_ROW_LENGTH             | bigint        | YES  |      | NULL        |       |
	| DATA_LENGTH                | bigint        | YES  |      | NULL        |       |
	| MAX_DATA_LENGTH            | bigint        | YES  |      | NULL        |       |
	| INDEX_LENGTH               | bigint        | YES  |      | NULL        |       |
	| DATA_FREE                  | bigint        | YES  |      | NULL        |       |
	| AUTO_INCREMENT             | bigint        | YES  |      | NULL        |       |
	| CREATE_TIME                | datetime      | YES  |      | NULL        |       |
	| UPDATE_TIME                | datetime      | YES  |      | NULL        |       |
	| CHECK_TIME                 | datetime      | YES  |      | NULL        |       |
	| TABLE_COLLATION            | varchar (32)   | YES  |      | utf8mb4_bin |       |
	| CHECKSUM                   | bigint        | YES  |      | NULL        |       |
	| CREATE_OPTIONS             | varchar (255)  | YES  |      | NULL        |       |
	| TABLE_COMMENT              | varchar (2048) | YES  |      | NULL        |       |
	| TIDB_TABLE_ID              | bigint        | YES  |      | NULL        |       |
	| TIDB_ROW_ID_SHARDING_INFO  | varchar (255)  | YES  |      | NULL        |       |
	| TIDB_PK_TYPE               | varchar (64)   | YES  |      | NULL        |       |
	| TIDB_PLACEMENT_POLICY_NAME | varchar (64)   | YES  |      | NULL        |       |
	| TIDB_TABLE_MODE            | varchar (16)   | YES  |      | NULL        |       |
	+----------------------------+---------------+------+------+-------------+-------+
	26 rows in set (0.00 sec)

    IMPORTANT: Don't put anything else in the response.

    EXAMPLE INPUT:
    Return 3 random SQL queries using this operation: or.

    EXAMPLE JSON OUTPUT:
    {
  "queries": [
    "CREATE DATABASE dbA DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
	CREATE DATABASE dbB CHARACTER SET latin1 COLLATE latin1_swedish_ci;
	CREATE TABLE dbA.t1 (id INT PRIMARY KEY, name VARCHAR(30) NOT NULL, status ENUM('a','b','c') DEFAULT 'a', created_at DATETIME);
	CREATE TABLE dbA.t2 (col1 INT, col2 VARCHAR(20), INDEX idx_col2(col2));
	CREATE TABLE dbB.x1 (x INT, y VARCHAR(10), INDEX idx_y(y));
	SELECT SCHEMA_NAME, DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE DEFAULT_CHARACTER_SET_NAME = 'utf8mb4' OR DEFAULT_COLLATION_NAME LIKE 'latin1%';
	SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_SCHEMA IN ('dbA','dbB') AND TABLE_NAME > 't1' ORDER BY TABLE_SCHEMA, TABLE_NAME;
	DROP TABLE dbB.x1; DROP TABLE dbA.t2; DROP TABLE dbA.t1; DROP DATABASE dbB; DROP DATABASE dbA;",
    "CREATE DATABASE MixedCase CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
	CREATE TABLE MixedCase.TCase (A INT, B VARCHAR(15), C DATETIME, INDEX idx_b(B));
	CREATE TABLE MixedCase.tcase2 (x INT PRIMARY KEY, y VARCHAR(5));
	SELECT SCHEMA_NAME, DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = 'MixedCase';
	SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_ROWS, AVG_ROW_LENGTH FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'MixedCase' AND TABLE_NAME >= 'TCase' ORDER BY TABLE_NAME DESC;
	DROP TABLE MixedCase.tcase2; DROP TABLE MixedCase.TCase; DROP DATABASE MixedCase;",
    "CREATE DATABASE dbNums CHARACTER SET utf8mb4 COLLATE utf8mb4_bin; CREATE TABLE dbNums.n1 (n INT PRIMARY KEY, v VARCHAR(10)); CREATE TABLE dbNums.n2 (n INT, v VARCHAR(10), INDEX idx_v(v)); CREATE TABLE dbNums.n3 (n INT, v VARCHAR(10), INDEX idx_v2(v(5))); SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME IN ('dbNums'); SELECT TABLE_SCHEMA, TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'dbNums' AND (TABLE_NAME = 'n1' OR TABLE_NAME LIKE 'n%') ORDER BY TABLE_NAME; DROP TABLE dbNums.n3; DROP TABLE dbNums.n2; DROP TABLE dbNums.n1; DROP DATABASE dbNums;",
    "CREATE DATABASE uni_db CHARACTER SET utf8 COLLATE utf8_unicode_ci; CREATE TABLE uni_db.u1 (id INT, txt NVARCHAR(20), PRIMARY KEY(id)); CREATE TABLE uni_db.u2 (id INT, val VARCHAR(20) CHARACTER SET latin1, INDEX idx_val(val)); SELECT SCHEMA_NAME, DEFAULT_CHARACTER_SET_NAME FROM information_schema.SCHEMATA WHERE DEFAULT_CHARACTER_SET_NAME <> 'latin1'; SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'uni_db' AND TABLE_NAME IN ('u1','u2') ORDER BY TABLE_NAME; DROP TABLE uni_db.u2; DROP TABLE uni_db.u1; DROP DATABASE uni_db;"
  ]
}`
	messages = append(messages, openai.SystemMessage(systemPrompt))

	userPromptTemplate := `Return %d random SQL queries using this operation: %s.`

	if len(existCases) > 0 {
		messages = append(messages, openai.UserMessage(fmt.Sprintf(userPromptTemplate, len(existCases), group)))

		existResponse := make([]string, 0, len(existCases))
		for _, c := range existCases {
			existResponse = append(existResponse, c.SQL)
		}
		assistantMessage, err := json.Marshal(simplePromptResponse{
			Queries: existResponse,
		})
		// should never happen
		if err != nil {
			logger.Global.Info("failed to marshal exist response", zap.Error(err))
			return nil
		}
		messages = append(messages, openai.AssistantMessage(string(assistantMessage)))
	}
	messages = append(messages, openai.UserMessage(fmt.Sprintf(userPromptTemplate, count, group)))

	return messages
}

// Unmarshal implements PromptGenerator.Unmarshal
func (g *infoschemaGenerator) Unmarshal(response string) []testcase.Case {
	var resp simplePromptResponse
	err := json.Unmarshal([]byte(response), &resp)
	if err != nil {
		logger.Global.Error("failed to unmarshal dml prompt response", zap.Error(err), zap.String("response", response))
		return nil
	}

	cases := make([]testcase.Case, 0, len(resp.Queries))
	for _, q := range resp.Queries {
		cases = append(cases, testcase.Case{
			SQL: q,
		})
	}

	return cases
}

func init() {
	registerPromptGenerator(&infoschemaGenerator{})
}
