# Bradley-Terry ELO Formulation and Statistical Methods

evalarena computes model skill ratings using the Bradley-Terry paired comparison model, solved via Maximum Likelihood Estimation (MLE), with non-parametric bootstrap resampling for confidence intervals and multi-objective Pareto frontier identification.

## 1. Bradley-Terry Preference Model

Let $K$ denote the number of competing models, indexed by $i, j \in \{1, \dots, K\}$. Each model $i$ has an unobserved latent skill rating $\beta_i \in \mathbb{R}$.

The probability that Model $i$ is preferred over Model $j$ in a pairwise match is modeled as:

$$P(i \succ j) = \frac{e^{\beta_i}}{e^{\beta_i} + e^{\beta_j}} = \frac{1}{1 + e^{-(\beta_i - \beta_j)}}$$

Converting the latent parameter $\beta_i$ to standard ELO scale:

$$\text{ELO}_i = 1000 + \frac{400}{\ln(10)} \cdot \beta_i$$

Under this scaling, a difference of 400 ELO points corresponds to a 10:1 win odds ratio:

$$P(i \succ j) = \frac{1}{1 + 10^{(\text{ELO}_j - \text{ELO}_i)/400}}$$

## 2. Maximum Likelihood Estimation (MLE)

Given observed tournament results:
- $W_{ij}$: number of times Model $i$ won against Model $j$.
- $T_{ij}$: number of ties between Model $i$ and Model $j$. Ties are treated as half a win for each model: $W'_{ij} = W_{ij} + 0.5 T_{ij}$.
- $N_{ij} = W_{ij} + W_{ji} + T_{ij}$: total matches between $i$ and $j$.

The log-likelihood of the observations given parameter vector $\boldsymbol{\beta} = (\beta_1, \dots, \beta_K)$ is:

$$\ln \mathcal{L}(\boldsymbol{\beta}) = \sum_{i < j} \left[ W'_{ij} \ln P(i \succ j) + W'_{ji} \ln P(j \succ i) \right] - \frac{\lambda}{2} \sum_{i=1}^K \beta_i^2$$

where $\lambda > 0$ is an L2 regularization parameter (default $\lambda = 10^{-4}$) to prevent numerical divergence when a model has zero defeats.

### 2.1 Newton-Raphson Optimization

The gradient with respect to $\beta_i$ is:

$$\frac{\partial \ln \mathcal{L}}{\partial \beta_i} = \sum_{j \ne i} \left( W'_{ij} - N_{ij} \frac{e^{\beta_i}}{e^{\beta_i} + e^{\beta_j}} \right) - \lambda \beta_i$$

The Hessian matrix entries are:

$$\frac{\partial^2 \ln \mathcal{L}}{\partial \beta_i^2} = -\sum_{j \ne i} N_{ij} \frac{e^{\beta_i} e^{\beta_j}}{(e^{\beta_i} + e^{\beta_j})^2} - \lambda$$

$$\frac{\partial^2 \ln \mathcal{L}}{\partial \beta_i \partial \beta_j} = N_{ij} \frac{e^{\beta_i} e^{\beta_j}}{(e^{\beta_i} + e^{\beta_j})^2} \quad (i \ne j)$$

We fix the baseline model skill $\sum_{i=1}^K \beta_i = 0$ or fix one reference model $\beta_{\text{ref}} = 0$. The parameter vector is updated iteratively:

$$\boldsymbol{\beta}^{(t+1)} = \boldsymbol{\beta}^{(t)} - \left[ \nabla^2 \ln \mathcal{L}(\boldsymbol{\beta}^{(t)}) \right]^{-1} \nabla \ln \mathcal{L}(\boldsymbol{\beta}^{(t)})$$

Iteration terminates when $||\boldsymbol{\beta}^{(t+1)} - \boldsymbol{\beta}^{(t)}||_\infty < 10^{-6}$ or maximum iterations (default 100) are reached.

## 3. Bootstrap Resampling for Confidence Intervals

To calculate empirical confidence intervals without relying on asymptotic normal approximations:

1. Let $\mathcal{D} = \{m_1, m_2, \dots, m_M\}$ denote the collection of all $M$ observed pairwise match outcomes.
2. For $b = 1, \dots, B$ (default $B = 1,000$):
   a. Sample $M$ match outcomes with replacement: $\mathcal{D}^{*(b)} \sim \text{Resample}(\mathcal{D})$.
   b. Fit the Bradley-Terry MLE model on $\mathcal{D}^{*(b)}$ to compute $\boldsymbol{\beta}^{*(b)}$ and corresponding $\text{ELO}_i^{*(b)}$.
3. For each model $i$, sort the bootstrap estimates: $\text{ELO}_i^{*(1)} \le \text{ELO}_i^{*(2)} \le \dots \le \text{ELO}_i^{*(B)}$.
4. The empirical 95% confidence interval is:

$$[\text{ELO}_{i, 2.5\%}, \text{ELO}_{i, 97.5\%}] = [\text{ELO}_i^{*(\lfloor 0.025 B \rfloor)}, \text{ELO}_i^{*(\lceil 0.975 B \rceil)}]$$

## 4. Pareto Efficiency Frontier

In enterprise production decisions, model selection requires balancing quality (ELO), inference cost (USD per 1M tokens), and generation latency (time to first token or total duration).

Let each evaluated configuration be characterized by:
$$\mathbf{x}_k = (\text{ELO}_k, C_k, L_k)$$
where:
- $\text{ELO}_k$: estimated Bradley-Terry rating (higher is better).
- $C_k$: blended cost in USD per 1M tokens (lower is better).
- $L_k$: median latency in milliseconds (lower is better).

### 4.1 Dominance Definition

Configuration $A$ strictly dominates configuration $B$ ($A \succ_{\text{Pareto}} B$) if and only if:
1. $\text{ELO}_A \ge \text{ELO}_B$, $C_A \le C_B$, and $L_A \le L_B$.
2. At least one inequality is strict:
   $$(\text{ELO}_A > \text{ELO}_B) \lor (C_A < C_B) \lor (L_A < L_B)$$

### 4.2 Frontier Determination

The Pareto Efficiency Frontier $\mathcal{P}$ is the subset of configurations that are not dominated by any other configuration in the evaluated set:

$$\mathcal{P} = \{ \mathbf{x}_k \in \mathcal{X} \mid \nexists \mathbf{x}_m \in \mathcal{X} \text{ such that } \mathbf{x}_m \succ_{\text{Pareto}} \mathbf{x}_k \}$$

Models lying on the frontier represent the rational choices for production deployment based on budget and latency constraints.
