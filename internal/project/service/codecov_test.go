// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/neglectedta/kres/internal/output/drone"
	"github.com/neglectedta/kres/internal/output/ghworkflow"
	"github.com/neglectedta/kres/internal/output/makefile"
	"github.com/neglectedta/kres/internal/project/service"
)

func TestCodeCovInterfaces(t *testing.T) {
	assert.Implements(t, (*makefile.Compiler)(nil), new(service.CodeCov))
	assert.Implements(t, (*drone.Compiler)(nil), new(service.CodeCov))
	assert.Implements(t, (*ghworkflow.Compiler)(nil), new(service.CodeCov))
}
