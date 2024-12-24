local image = 'zachfi/shell:latest';


local pipeline(name) = {
  kind: 'pipeline',
  name: name,
  steps: [],
  depends_on: [],
  trigger: {
    ref: [
      'refs/heads/main',
      'refs/heads/dependabot/**',
      'refs/pull/*/head',
    ],
  },
};

local withPipelineOnlyTags() = {
  trigger+: {
    ref: [
      'refs/tags/v*',
    ],
  },
};

local step(name) = {
  name: name,
  image: 'zachfi/build-image',
  pull: 'always',
  commands: [],
};

local make(target) = step(target) {
  commands: ['make %s' % target],
};


local withTags() = {
  when+: {
    ref+: [
      'refs/tags/v*',
    ],
  },
};

[
  (
    pipeline('lint') {
      steps: [
        make('lint'),
      ],
    }
  ),
]
